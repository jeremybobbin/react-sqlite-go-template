### build

```
./configure
make
```

### run

```
make run
```

### watch

```
./watch.sh
```

### about

this template shows how one can use make with esbuild to manage project dependencies, as an alternative to webpack.

---

#### build time (npm warnings redacted)

```
$ time sh -c './configure && make -j'
go build
sqlite3 sql.db < sql.sql
esbuild  App.jsx > App.mjs
esbuild  Todo.jsx > Todo.mjs
npm i
touch sql.db
touch -c node_modules
esbuild --sourcemap --bundle App.mjs > App.js

real    0m1.099s
user    0m1.814s
sys     0m0.432s
```
