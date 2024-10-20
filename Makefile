ENV_DIR="./configs/envs"
ENV_SUFFIX=".env.example"


.PHONY: dev-start
dev-start:
	@ docker compose up --remove-orphans --build
	@ docker compose down --remove-orphans

.PHONY: env
env:
	@ eval ls "${ENV_DIR}/*${ENV_SUFFIX}" \
		| xargs -I {} basename --suffix "${ENV_SUFFIX}" {} \
		| xargs -I {} cp --update=none "${ENV_DIR}/{}${ENV_SUFFIX}" "${ENV_DIR}/{}.env"
