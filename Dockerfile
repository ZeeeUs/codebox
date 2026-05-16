FROM golang:1.25-bookworm

ENV CGO_ENABLED=1
ENV PATH="/root/go/bin:${PATH}"

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    build-essential \
    ca-certificates \
    cargo \
    curl \
    git \
    gdb \
    less \
    nodejs \
    npm \
    procps \
    rustc \
    && rm -rf /var/lib/apt/lists/*

RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

RUN npm i -g @openai/codex@latest \
    && codex --version

WORKDIR /workspace

CMD ["bash"]
