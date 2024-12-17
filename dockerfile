FROM shivam010/golang AS build

WORKDIR /app
COPY . .

RUN go build -o main main.go

FROM alpine

WORKDIR /app
COPY --from=build /app/main .

CMD ["./main"]
