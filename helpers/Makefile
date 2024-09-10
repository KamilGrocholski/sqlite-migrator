SHELL:=bash

SQLITE=sqlite3
DIR=./migrations
DB=test.db
EXT_UP=up.sql
EXT_DOWN=down.sql
SCHEMA=./schema.sql
FILES_UP=$(wildcard $(DIR)/.$(EXP_UP))
FILES_DOWN=$(wildcard $(DIR)/.$(EXP_DOWN))

DB_VERSION=$$($(SQLITE) $(DB) -cmd "pragma user_version;" -echo ".quit")
MIGRATIONS_COUNT=$$(find $(DIR) -name "*.$(EXT_UP)" | wc -l)

help:

.PHONY=db-version
db-version: ## get sqlite 'user_version' and quit from $(SQLITE)
	sh -c '$(SQLITE) $(DB) -cmd "pragma user_version;" -echo ".quit"'

.PHONY=db-rm
db-rm: 
	rm $(DB)

.PHONY=db-push
db-push:
	@if [ "$(DB_VERSION)" = "$(MIGRATIONS_COUNT)" ]; then \
		echo "nothing new to push"; \
	else \
		rm -f $(SCHEMA); \
		find $(DIR) -name "*.$(EXT_UP)" | sort | xargs cat >> $(SCHEMA); \
		$(SQLITE) -echo $(DB) < $(SCHEMA); \
		sh -c '$(SQLITE) $(DB) -cmd "pragma user_version=$(MIGRATIONS_COUNT);" -echo ".quit"'; \
	fi

.PHONY=db-migrate-up
db-migrate-up: build-schema-uped
	$(SQLITE) -echo $(DB) < $(SCHEMA)

.PHONY=db-migrate-down
db-migrate-down: build-schema-downed
	$(SQLITE) -echo $(DB) < $(SCHEMA)

.PHONY=db-new-migrate
db-new-migrate: 
	@NOW=$$(date +%s); \
		echo $$NOW \
			| xargs -I {} \
				sh -c "touch $(DIR)/{}_$(desc).$(EXT_UP) & touch $(DIR)/{}_$(desc).$(EXT_DOWN)"

.PHONY=build-schema-uped
build-schema-uped:
	rm -f $(SCHEMA)
	find $(DIR) -name "*.$(EXT_UP)" | sort | xargs cat >> $(SCHEMA)

.PHONY=build-schema-downed
build-schema-downed:
	rm $(SCHEMA)
	find $(DIR) -name "*.$(EXT_DOWN)" | sort -r | xargs cat >> $(SCHEMA)
	
.PHONY=count-migrations
count-migrations:
	find $(DIR) -name "*.$(EXT_UP)" | wc -l
