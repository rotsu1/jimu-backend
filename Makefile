test:
	docker-compose -f docker-compose.test.yml down --volumes --remove-orphans || true
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit