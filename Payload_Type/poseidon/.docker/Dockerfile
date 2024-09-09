FROM itsafeaturemythic/mythic_go_macos:latest

WORKDIR /Mythic/

COPY [".", "."]

RUN make build

RUN apt-get install g++-x86-64-linux-gnu libc6-dev-amd64-cross -y

CMD make run