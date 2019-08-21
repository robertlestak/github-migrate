dist: clean
	mkdir -p dist
	go build -o dist/ghmigrate cmd/*

clean:
	rm -rf dist

.PHONY: dist clean
