.PHONY: build ui backend containerapps run run-containerapps clean

build: ui backend containerapps

ui:
	cd ui && pnpm build && pnpm generate

backend:
	cd deployberry && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -buildvcs=false -o ../bin/deployberry-linux-amd64 .

containerapps:
	cd deployberry && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags containerapps_only -trimpath -ldflags="-s -w" -buildvcs=false -o ../bin/containerapps-linux-amd64 .

run:
	cd deployberry && go run .

run-containerapps:
	cd deployberry && go run -tags containerapps_only .

clean:
	rm -rf bin/ releases/ ui/dist
	mkdir -p ui/dist
	echo "Placeholder" > ui/dist/placeholder.txt
