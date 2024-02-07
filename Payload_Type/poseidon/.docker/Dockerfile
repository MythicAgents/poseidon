FROM itsafeaturemythic/mythic_go_macos:latest

WORKDIR /Mythic/

COPY [".", "."]

RUN make build

CMD make run