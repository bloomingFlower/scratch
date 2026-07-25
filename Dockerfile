# Multi-stage: only the final stage's layers are published, so the toolchain and
# the source tree never reach the registry. (The previous version declared
# `AS builder` but had no second stage, so the published image WAS the builder —
# it carried .env with a live DB password, the .git history and id_rsa.enc, and
# weighed 1.6GB.)
#
# The builder runs on the native build platform and cross-compiles to the target
# arch (pure Go, CGO disabled), avoiding a QEMU-emulated build.
# Pinned to the toolchain this vendored tree was built with: the vendored
# protobuf v1.33.0 does not compile under Go 1.24 (its build-tagged unsafe
# implementation files don't match, giving "undefined: value/nilType" errors).
FROM --platform=$BUILDPLATFORM golang:1.21.3 AS builder

LABEL maintainer="JYY <yourrubber@duck.com>"

WORKDIR /app

ARG TARGETOS TARGETARCH

# Build in module mode (-mod=mod), NOT against the committed vendor/ tree.
# That tree is missing protobuf's build-tagged unsafe implementation
# (only value_pure.go is vendored), so `go build -mod=vendor` fails with
# "undefined: value/nilType" — i.e. the vendored path no longer builds at all,
# regardless of this change. go.sum still pins every dependency.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=mod -a -installsuffix cgo -o main .

# --- runtime ---
FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && addgroup -g 10001 appuser \
    && adduser -D -u 10001 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/main .
# handler_view.go renders html/view.html at runtime.
COPY --from=builder /app/html ./html

USER appuser

EXPOSE 50051

CMD ["./main"]
