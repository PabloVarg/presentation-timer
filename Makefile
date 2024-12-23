ENV_DIR="configs/envs"
ENV_SUFFIX=".env.example"
MIGRATIONS_DIR="migrations/"

include configs/envs/db.env


.PHONY: dev-start
dev-start:
	@ docker compose up --remove-orphans --build
	@ docker compose down --remove-orphans

.PHONY: env
env:
	@ eval ls "${ENV_DIR}/*${ENV_SUFFIX}" \
		| xargs -I {} basename --suffix "${ENV_SUFFIX}" {} \
		| xargs -I {} cp --update=none "${ENV_DIR}/{}${ENV_SUFFIX}" "${ENV_DIR}/{}.env"

.PHONY: migrations-up
migrations-up:
	@ goose \
		--dir "${MIGRATIONS_DIR}" \
		postgres "user=${POSTGRES_USER} dbname=${POSTGRES_DB} sslmode=${POSTGRES_SSLMODE} password=${POSTGRES_PASSWORD} host=127.0.0.1" \
		up

.PHONY: migrations-down-last
migrations-down-last:
	@ goose \
		--dir "${MIGRATIONS_DIR}" \
		postgres "user=${POSTGRES_USER} dbname=${POSTGRES_DB} sslmode=${POSTGRES_SSLMODE} password=${POSTGRES_PASSWORD} host=127.0.0.1" \
		down 1
