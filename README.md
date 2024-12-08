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
cat sql.sql | sqlite3 sql.db
esbuild  App.jsx > App.mjs
esbuild  Todo.jsx > Todo.mjs
npm i
touch -c node_modules
esbuild --sourcemap --bundle App.mjs > public/index.js

real    0m1.948s
user    0m3.135s
sys     0m0.542s
```
