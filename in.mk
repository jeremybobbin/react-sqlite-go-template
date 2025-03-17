.PHONY: build run serve clean nuke
.SUFFIXES: .jsx .mjs .js .sql .db

.jsx.mjs:
	$(ESC) $(ESCFLAGS) $< > $@ || (rm -f $@; exit 1)

.mjs.js:
	$(ESL) $(ESLFLAGS) $< > $@ || (rm -f $@; exit 1)

.sql.db:
	$(SQL) $@ < $< && touch $@

build: $(BIN) sql.db App.js

run: build
	./$(BIN)

serve:
	./$(BIN)

view: build
	chromium --no-default-browser-check \
		--user-data-dir=./.cache  \
		--enable-logging=stderr --v=1 \
		"http://localhost:8080/" 2>&1 | \
			 ./debug/chromium-log

$(BIN): $(GO_FILES)
	go build
	touch $(BIN)

node_modules: package.json
	$(NPM) i && touch -c node_modules

App.js: App.mjs Todo.mjs node_modules

clean:
	rm -f *.js *.mjs
	go clean

nuke: clean
	rm -rf sql.db node_modules Makefile
