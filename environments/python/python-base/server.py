#!/usr/bin/env python
import importlib
import json
import logging
import os
import threading
import sys
import time

from flask import Flask, request, abort, g
from gevent.pywsgi import WSGIServer
from prometheus_client import PrometheusForFission

IS_PY2 = (sys.version_info.major == 2)

PATH_CONFIGS = "/configs"
PATH_SECRETS = "/secrets"
PUSHGATEWAY_URL_DEFAULT = "fh-prometheus-pushgateway.fission:9091"  # may be overwritten by configs


def import_src(path):
    if IS_PY2:
        import imp
        return imp.load_source('mod', path)
    else:
        # the imp module is deprecated in Python3. use importlib instead.
        return importlib.machinery.SourceFileLoader('mod', path).load_module()


def synchronized(func):
    """
    a simple function lock
    """
    func.__lock__ = threading.Lock()

    def lock_func(*args, **kwargs):
        with func.__lock__:
            return func(*args, **kwargs)

    return lock_func


class Info(object):
    """
    for Cache class
    """

    def __init__(self, value, timeout):
        """
        存储的内容和过期的时间
        :param value:
        :param timeout:
        """
        self.value = value
        self.timeout = timeout


class Cache(object):
    """cache"""

    def __init__(self):
        self.content = dict()  # key: Info

    @synchronized
    def put(self, key, func=lambda x: x, param=None, timeout=0, use_old=False, old_timeout=30):
        """
        存放信息
        :param key: 键
        :param func: 获取值的函数
        :param param: func的参数
        :param timeout: 过期时间，单位秒，0表示永不过期
        :param use_old: 当数据超时且获取新值的函数没有成功，是否返回超时了的旧数据
        :param old_timeout: 超时的数据可以继续存活的时间
        :return:
        """
        now = time.time()
        timeout += now if timeout != 0 else 0
        old_timeout += now
        value = self.get(key, no_delete=use_old)
        if value is not None:
            return value
        value = func(param)
        if value is None and use_old:
            if key in self.content:
                info = self.content[key]
                if info.timeout != 0:
                    info.timeout = old_timeout
                return info.value
        else:
            self.content[key] = Info(value, timeout)
        return value

    @synchronized
    def pop(self, key):
        if key in self.content:
            self.content.pop(key)

    def get(self, key, no_delete=False):
        """
        读取信息
        :param key: 键
        :param no_delete: 不执行删除操作
        """
        if key not in self.content:
            return None
        info = self.content[key]
        if info.timeout > time.time() or info.timeout == 0:
            return info.value
        else:
            if no_delete is False:
                self.pop(key)
            return None

    def get_and_write(self, key: str, func=lambda x: x, param=None, timeout=0, use_old=True, old_timeout=30):
        ans = self.get(key, no_delete=True)
        if ans is not None:
            return ans
        return self.put(key, func, param, timeout, use_old, old_timeout)


def add_params(con, path, key, value):
    """
    在字典中添加内容，path 是字典的路径，key是key
    """
    pos = con
    for p in path:
        if p not in pos:
            pos[p] = {}
        pos = pos[p]
    pos[key] = value


def read_config(base_dir):
    """读取目录下的配置文件"""
    g = os.walk(base_dir)
    config = dict()
    for current_path, dir_list, file_list in g:  # BFS
        for file_name in file_list:
            paths = current_path.split("/")[2:]  # 第一个是空，第二个是config或者secrets
            value = open(os.path.join(current_path, file_name)).read()  # 读取文件中的内容
            add_params(config, paths, file_name, value)
    return config


class FuncApp(Flask):
    def __init__(self, name, loglevel=logging.DEBUG):
        super(FuncApp, self).__init__(name)

        # init the class members
        self.userfunc = None
        self.metric_handler = None
        self.configs = {}
        self.secrets = {}
        self.cache = Cache()
        self.root = logging.getLogger()
        self.ch = logging.StreamHandler(sys.stdout)

        #
        # Logging setup.  TODO: Loglevel hard-coded for now. We could allow
        # functions/routes to override this somehow; or we could create
        # separate dev vs. prod environments.
        #
        self.root.setLevel(loglevel)
        self.ch.setLevel(loglevel)
        self.ch.setFormatter(logging.Formatter('[%(levelname)s]- %(message)s'))

        # self.logger.addHandler(self.ch)

        #
        # Register the routers
        #
        @self.route('/specialize', methods=['POST'])
        def load():
            self.logger.info('env_info: /specialize called')
            # load user function from codepath
            self.userfunc = import_src('/userfunc/user').main
            return ""

        @self.route('/v2/specialize', methods=['POST'])
        def loadv2():
            body = request.get_json()
            filepath = body['filepath']
            handler = body['functionName']
            self.logger.info(
                'env_info: /v2/specialize called with  filepath = "{}"   handler = "{}"'.format(filepath, handler))

            # handler looks like `path.to.module.function`
            parts = handler.rsplit(".", 1)
            if len(handler) == 0:
                # default to main.main if entrypoint wasn't provided
                moduleName = 'main'
                funcName = 'main'
            elif len(parts) == 1:
                moduleName = 'main'
                funcName = parts[0]
            else:
                moduleName = parts[0]
                funcName = parts[1]
            self.logger.debug('env_info: moduleName = "{}"    funcName = "{}"'.format(moduleName, funcName))

            # check whether the destination is a directory or a file
            if os.path.isdir(filepath):
                # add package directory path into module search path
                sys.path.append(filepath)

                self.logger.debug('env_info: __package__ = "{}"'.format(__package__))
                if __package__:
                    mod = importlib.import_module(moduleName, __package__)
                else:
                    mod = importlib.import_module(moduleName)

            else:
                # load source from destination python file
                mod = import_src(filepath)

            # load user function from module
            self.userfunc = getattr(mod, funcName)

            # set configs and secrets
            self.configs = read_config(PATH_CONFIGS)
            self.secrets = read_config(PATH_SECRETS)
            self.logger.info("configs: {}".format(json.dumps(self.configs, indent=2)))
            pushgateway_url_temp = self.configs.get("fission-secret-configmap", {}).get("fission-function-global-configmap", {}).get("url-pushgateway", None)
            if pushgateway_url_temp is not None:
                pushgateway_url = pushgateway_url_temp
            else:
                pushgateway_url = PUSHGATEWAY_URL_DEFAULT

            # set metric_handler
            fns = body.get('FunctionMetadata', {}).get('namespace', "unknown")
            fn = body.get('FunctionMetadata', {}).get('name', "unknown")
            prefix = fns + "-" + fn
            update_time = body.get('FunctionMetadata', {}).get('managedFields', [])
            if len(update_time) != 0:
                update_time = update_time[0].get('time', None)
            else:
                update_time = body.get('FunctionMetadata', {}).get('creationTimestamp', "unknown")
            self.metric_handler = PrometheusForFission(prefix, update_time, pushgateway_url, self.logger)

            return ""

        @self.route('/healthz', methods=['GET'])
        def healthz():
            return "", 200

        @self.route('/', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE'])
        def f():
            if self.userfunc is None:
                print("Generic container: no requests supported")
                abort(500)
            # use g to pass parameter to function
            g.logger = self.logger
            g.metric_handler = self.metric_handler
            g.configs = self.configs  # 函数配置的参数
            g.secrets = self.secrets
            g.cache = self.cache
            res = self.userfunc()
            return res


app = FuncApp(__name__, logging.DEBUG)

#
# TODO: this starts the built-in server, which isn't the most
# efficient.  We should use something better.
#

app.logger.info("env_info: Starting gevent based server")
svc = WSGIServer(('0.0.0.0', 8888), app)
svc.serve_forever()
