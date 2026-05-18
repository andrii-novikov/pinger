-include .env
export

RASPBERRY_PATH ?= /home/pi

build:
	GOOS=linux GOARCH=arm go build -o bin/raspbery/pinger .
	go build -o bin/mac/pinger .

deploy: build
	scp bin/raspbery/pinger $(RASPBERRY_HOST):$(RASPBERRY_PATH)/pinger

enable-cron:
	ssh $(RASPBERRY_HOST) '(crontab -l 2>/dev/null | grep -v "$(RASPBERRY_PATH)/pinger"; echo "$(CRON_SCHEDULE) TELEGRAM_TOKEN=$(TELEGRAM_TOKEN) TELEGRAM_CHAT_ID=$(TELEGRAM_CHAT_ID) $(RASPBERRY_PATH)/pinger") | crontab -'
