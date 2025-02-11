.PHONY: build run serve clean nuke
.SUFFIXES: .jsx .mjs .js .sql .db

.jsx.mjs:
	$(ESC) $(ESCFLAGS) $< > $@

.mjs.js:
	$(ESL) $(ESLFLAGS) $< > $@

.sql.db:
	$(SQL) $@ < $<
	touch $@

build: $(BIN) sql.db App.js

run: build
	./$(BIN)

serve:
	./$(BIN)

$(BIN): $(GO_FILES)
	go build

node_modules: package.json
	$(NPM) i
	touch -c node_modules

App.js: App.mjs Todo.mjs node_modules

clean:
	rm -f *.js *.mjs
	go clean

nuke: clean
	rm -rf sql.db node_modules
