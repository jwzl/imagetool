.PHONY:	binary

binary:
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	go build imagetool.go 
	@chmod 777 imagetool


.PHONY: clean
clean:
	@rm -rf imagetool
	@echo "[clean Done]"
