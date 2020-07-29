FROM alpine

WORKDIR /app

ENV TZ Asia/Shanghai

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --update gcc && \
    apk upgrade --no-cache && \
    apk add --no-cache python3 python3-dev build-base libev-dev libffi-dev bash py-pip git && \
    pip3 install --upgrade pip && \
    rm -r /root/.cache

COPY requirements.txt /app
RUN pip3 install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple/

# todo 切换仓库 更新prometheus_client_python
RUN cd /tmp && git clone https://git.jingtao.fun/jingtao/prometheus_client_python.git && \
    cd prometheus_client_python && git pull && python3 setup.py install && echo "2"

COPY . /app

ENTRYPOINT ["python3"]
CMD ["server.py"]