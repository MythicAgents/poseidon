FROM itsafeaturemythic/xgolang_payload:0.1.2

# Download Garble
RUN go install mvdan.cc/garble@latest

# Needed package for Windows cross compiling
RUN apt update; apt-get -y install gcc-mingw-w64