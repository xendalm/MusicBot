.PHONY: compose-up
compose-up:
	docker-compose build
	docker-compose up -d

.PHONY: compose-down
compose-down:
	docker-compose down