FROM node:22.22.3-bookworm-slim
ARG CODEX_VERSION=latest
RUN apt-get update && apt-get install -y --no-install-recommends bash ca-certificates curl git less procps \
    && rm -rf /var/lib/apt/lists/*
RUN npm i -g "@openai/codex@${CODEX_VERSION}" && codex --version
ENV CGO_ENABLED=1
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ARG INSTALL_GO=false
RUN if [ "$INSTALL_GO" = "true" ]; then \
      apt-get update && apt-get install -y --no-install-recommends build-essential curl gdb \
      && rm -rf /var/lib/apt/lists/* \
      && curl -fsSL "https://go.dev/dl/go1.26.5.linux-$(dpkg --print-architecture).tar.gz" \
        | tar -C /usr/local -xz \
      && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
      && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt \
      && go version \
      && go install golang.org/x/tools/gopls@v0.23.0 \
      && go install github.com/go-delve/delve/cmd/dlv@v1.26.1 \
      && go install honnef.co/go/tools/cmd/staticcheck@v0.7.0 \
      && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4; \
    fi
ARG INSTALL_RUST=false
RUN if [ "$INSTALL_RUST" = "true" ]; then \
      apt-get update && apt-get install -y --no-install-recommends cargo rustc \
      && rm -rf /var/lib/apt/lists/*; \
    fi
ARG LANGUAGE_SET=""
LABEL io.codebox.runtime-version="6" io.codebox.languages="${LANGUAGE_SET}"
WORKDIR /workspace
CMD ["bash"]
