# Multi-stage: only the final stage's layers are published, so the toolchain and
# the source tree never reach the registry. (The previous version declared
# `AS builder` but had no second stage, so the published image WAS the builder —
# it carried .env with a live DB password, the .git history and id_rsa.enc, and
# weighed 1.6GB.)
#
# The builder runs on the native build platform and cross-compiles to the target
# arch (pure Go, CGO disabled), avoiding a QEMU-emulated build.
# Use a patched toolchain and the module versions pinned in go.mod/go.sum.
FROM --platform=$BUILDPLATFORM golang:1.26.11 AS builder

LABEL maintainer="JYY <yourrubber@duck.com>"

WORKDIR /app

ARG TARGETOS TARGETARCH

# Build in module mode: the historical committed vendor tree is incomplete.
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
