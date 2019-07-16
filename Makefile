GO := go

intgenerator: main.go
	$(GO) build

clean:
	rm intgenerator