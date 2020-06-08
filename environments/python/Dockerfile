FROM alpine

WORKDIR /app

RUN apk add --no-cache python3 python3-dev py-pip build-base libev-dev libffi-dev jpeg-dev zlib-dev && \
    pip3 install --upgrade pip && \
    rm -r /root/.cache

ENV LIBRARY_PATH=/lib:/usr/lib

COPY requirements.txt /app
RUN pip3 install -r requirements.txt

COPY . /app

ENTRYPOINT ["python3"]
CMD ["server.py"]
