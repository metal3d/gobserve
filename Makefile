
GETVERSION:=cat .goxc.json | python -c 'import json,sys;o=json.load(sys.stdin);print o["PackageVersion"]' 
VERSION:=$(shell $(GETVERSION))
FPM=~/bin/fpm

version:
	@echo $(VERSION)


newtag:
	goxc bump
	git tag -a `$(GETVERSION)`

all: goxc rpm

goxc:
	goxc

rpm: cleantmp
	mkdir .tmp
	cd .tmp && \
		mkdir -p usr/bin usr/share/docs/gobserve/ && \
		tar zxf ../build/$(VERSION)/gobserve_$(VERSION)_linux_amd64.tar.gz &&\
		mv gobserve_$(VERSION)_linux_amd64/gobserve usr/bin/ &&\
		mv gobserve_$(VERSION)_linux_amd64/README.md usr/share/docs/gobserve/ &&\
		$(FPM) -e --description "Automatic command launcher on file system changes" --license BSD --vendor "Patrice FERLET <metal3d@gmail.com>" --rpm-changelog ../CHANGELOG -v $(VERSION) -n gobserve -s dir -t rpm usr
	cp .tmp/*.rpm build/0.0.1/
	$(MAKE) cleantmp

cleantmp:
	rm -rf .tmp


clean: cleantpm
	rm -rf build
