FROM heroiclabs/nakama-pluginbuilder:3.38.0 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /backend
COPY . .
RUN go mod tidy
RUN go build --trimpath --mod=mod --buildmode=plugin -o ./backend.so

FROM heroiclabs/nakama:3.38.0

COPY --from=builder /backend/backend.so /nakama/data/modules/
