.PHONY: build run serve clean nuke
.SUFFIXES: .jsx .mjs .js .sql .db

.jsx.mjs:
	$(ESC) $(ESCFLAGS) $< > $@ || (rm -f $@; exit 1)

.mjs.js:
	$(ESL) $(ESLFLAGS) $< > $@ || (rm -f $@; exit 1)

.sql.db:
	$(SQL) $@ < $< && touch $@

build: $(BIN) sql.db App.js $(BIN).service

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

install: build
	mkdir -p $(PREFIX)/lib/systemd/system/
	cp $(BIN) $(PREFIX)/bin/$(BIN)
	mkdir -p $(PREFIX)/lib/systemd/system/
	cp $(BIN).service $(PREFIX)/lib/systemd/system/$(BIN).service

$(BIN).service: in.service
	sed 's,@PORT,$(PORT),g; s,@PREFIX,$(PREFIX),g; s,@BIN,$(BIN),g' in.service > $@

$(BIN): $(GO_FILES)
	go build
	touch $(BIN)

package.json:
	$(NPM) i $(NPM_PACKAGES) && touch -c $@

App.js: App.mjs Todo.mjs package.json

clean:
	rm -f *.js *.mjs ???*.service
	go clean

nuke: clean
	rm -rf sql.db package.json package-lock.json node_modules Makefile
