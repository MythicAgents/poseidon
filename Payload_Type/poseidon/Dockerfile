FROM itsafeaturemythic/mythic_go_macos:latest

ARG TARGETARCH

RUN chmod +x /go/bin/garble
RUN garble version

WORKDIR /

RUN if [ "$TARGETARCH" = "arm64" ]; then \
    echo "Installing ARM-specific packages" && \
    # Add your ARM-specific commands here; \
    wget https://github.com/indygreg/apple-platform-rs/releases/download/apple-codesign%2F0.29.0/apple-codesign-0.29.0-aarch64-unknown-linux-musl.tar.gz && \
    gunzip apple-codesign-0.29.0-aarch64-unknown-linux-musl.tar.gz && \
    tar -xvf apple-codesign-0.29.0-aarch64-unknown-linux-musl.tar && \
    mv apple-codesign-0.29.0-aarch64-unknown-linux-musl/rcodesign /rcodesign && \
    chmod +x /rcodesign; \
    elif [ "$TARGETARCH" = "amd64" ]; then \
    echo "Installing AMD-specific packages" && \
    # Add your AMD-specific commands here; \
    wget -o /rcodesign.tar.gz https://github.com/indygreg/apple-platform-rs/releases/download/apple-codesign%2F0.29.0/apple-codesign-0.29.0-x86_64-unknown-linux-musl.tar.gz && \
    gunzip apple-codesign-0.29.0-x86_64-unknown-linux-musl.tar.gz && \
    tar -xvf apple-codesign-0.29.0-x86_64-unknown-linux-musl.tar && \
    mv apple-codesign-0.29.0-x86_64-unknown-linux-musl/rcodesign /rcodesign && \
    chmod +x /rcodesign; \
    fi

WORKDIR /Mythic/

COPY [".", "."]

RUN make build

RUN apt-get install g++-x86-64-linux-gnu libc6-dev-amd64-cross -y

CMD make run