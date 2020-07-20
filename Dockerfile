FROM golang:1.14 AS builder

#### Set Go environment
# Disable CGO to create a self-contained executable
# Do not enable unless it's strictly necessary
ENV CGO_ENABLED 0
# Set Linux as target
ENV GOOS linux

### Prepare build image
RUN apt-get update && apt-get install -y upx-ucl zip ca-certificates tzdata
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
RUN useradd --home /app/ -M appuser

WORKDIR /src/

### Copy Go modules files and cache dependencies
# If dependencies do not changes, these two lines are cached (speed up the build)
COPY go.* ./
RUN go mod download

### Copy Go code
COPY cmd cmd
COPY service service

### Set some build variables
ARG APP_VERSION
ARG BUILD_DATE

### Build bot, strip debug symbols and compress with UPX
WORKDIR /src/cmd/bot/
RUN go build -mod=readonly -ldflags "-extldflags \"-static\" -X main.APP_VERSION=${APP_VERSION} -X main.BUILD_DATE=${BUILD_DATE}" -a -installsuffix cgo -o goworker .
RUN strip goworker
RUN upx -9 goworker


### Create final container from scratch
FROM scratch

### Populate scratch with CA certificates and Timezone infos from the builder image
ENV ZONEINFO /zoneinfo.zip
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /zoneinfo.zip /
COPY --from=builder /etc/passwd /etc/passwd

### Copy the build executable from the builder image
WORKDIR /app/
COPY --from=builder /src/cmd/bot/goworker .

### Set some build variables
ARG APP_VERSION
ARG BUILD_DATE
ARG PROJECT_NAME
ARG GROUP_NAME

### Downgrade to user level (from root)
USER appuser

### Executable command
CMD ["/app/goworker"]
