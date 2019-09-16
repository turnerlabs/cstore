FROM golang:1.13.7

RUN apt-get update && apt-get -y upgrade && \
    apt-get install --no-install-recommends -y python-pip && \
    rm -rf /var/lib/apt/lists/* && \
    curl -O https://bootstrap.pypa.io/get-pip.py && \
    python get-pip.py --user && \
    python -m pip install awscli && \
    go get github.com/mitchellh/gox && \
    go get github.com/tcnksm/ghr

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]