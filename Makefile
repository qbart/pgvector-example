.PHONY: dev
dev:
	DATABASE_URL=postgres://app:secret@localhost:5432/dev?sslmode=disable \
	air

.PHONY: gen
gen:
	templ generate

.PHONY: mig
mig:
	goose -dir db/migrations create ${NAME} sql

.PHONY: deps
deps:
	go mod download

.PHONY: install
install:
	go install github.com/cosmtrek/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: css
css:
	bun run tailwindcss -i ./assets/css/style.css -o ./ui/style.css --watch

.PHONY: changelog
changelog:
	git-chglog -o CHANGELOG.md --next-tag ${TAG}
