FROM itsafeaturemythic/mythic_go_macos:latest

RUN ARCH=$(dpkg --print-architecture) && \
    wget -q https://go.dev/dl/go1.26.4.linux-${ARCH}.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.26.4.linux-${ARCH}.tar.gz && \
    rm go1.26.4.linux-${ARCH}.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOROOT="/usr/local/go"

RUN chmod +x /go/bin/garble && garble version

RUN apt-get install g++-x86-64-linux-gnu libc6-dev-amd64-cross -y

WORKDIR /Mythic/

# dependencies layer: only rebuilt when go.mod/go.sum change
COPY go.mod go.sum ./
COPY pikka/agent_code/go.mod pikka/agent_code/go.sum ./pikka/agent_code/
RUN go mod download && \
    cd pikka/agent_code && go mod download

COPY [".", "."]
RUN go build -o main . && cp main /

CMD make run
