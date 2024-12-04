.POSIX:
.PHONY: build run clean nuke
.SUFFIXES: .m4 .sql .db .js .jsx .mjs

include config.mk

.jsx.mjs:
	$(ESC) $(ESCFLAGS) $< > $@

.sql.db:
	cat $< | $(SQL) $@

build: $(BIN) sql.db ./public/index.js

run: build
	./$(BIN)

serve:
	./$(BIN)

$(BIN): $(GO_FILES)
	go build

node_modules: package.json
	$(NPM) i
	touch -c node_modules

public/index.js: App.mjs Todo.mjs node_modules
	$(ESL) $(ESLFLAGS) $< > $@

config.mk:
	./configure

clean:
	rm -f public/index.js *.mjs
	go clean

nuke: clean
	rm -rf sql.db node_modules
