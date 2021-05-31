all: echo .echo

echo: echo.go
	go build -o echo echo.go

.echo: echo.go Dockerfile
	go build -o /dev/null echo.go # quick fail
	docker build -t duglin/echo .
	docker push duglin/echo
	touch .echo

icr: .icr

.icr: echo.go
	go build -o /dev/null echo.go
	docker build -t us.icr.io/dugs/echo .
	docker push us.icr.io/dugs/echo
	-kn service delete echo > /dev/null 2&>1
	kn service create echo --image us.icr.io/dugs/echo
	curl $$(kn service describe echo -o go-template="{{.status.url}}")
	touch .icr

clean:
	rm -f .echo echo
