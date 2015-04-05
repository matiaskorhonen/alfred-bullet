.PHONY: build clean extension

build:
	mkdir -p ./build
	go build -o ./build/alfred-bullet

workflow: build
	mkdir -p ./tmp
	rm -rf ./tmp/AlfredBullet
	cp -Rp ./workflow ./tmp/workflow
	cp -fp ./build/alfred-bullet ./tmp/workflow/alfred-bullet
	perl -ne 's/README_INSERTION/`cat README.md`/e;print' ./workflow/info.plist > ./tmp/workflow/info.plist
	cd ./tmp/workflow && defaults write Info.plist "readme" -string 'Test'
	cd ./tmp/workflow && zip -r AlfredBullet.alfredworkflow *
	mv ./tmp/workflow/AlfredBullet.alfredworkflow ./build/AlfredBullet.alfredworkflow

clean:
	rm -rf ./tmp
	rm -rf ./build
	go clean
