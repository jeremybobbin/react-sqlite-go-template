#!/bin/sh

deps=$(mktemp)

trap 'if [ -n "$pid" ]; then kill "$pid"; fi; exit 0' INT

while
	make -j build
	make serve & pid=$!
	make -p |
		awk '
			/^.PHONY:/ {
				for (i=2; i <= NF; i++) {
					phony[$((i))];
				}
			}
			!/^[.%#]/ && !/^\(%\)/ && /^\S+*:/ && !/^make:/ && NF > 0 {
				for (i=2; i <= NF; i++) {
					files[$((i))];
				}
			}
			END {
				for (file in files) {
					if (file in phony || file == "sql.db") {
						continue
					}
					print file
				}
			}
		' | inotifywait -q --fromfile=- -e CREATE,MOVED_FROM,DELETE,MODIFY,ATTRIB
	kill "$pid"
	wait
do :; done
