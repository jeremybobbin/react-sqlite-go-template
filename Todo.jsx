import React from 'react';
import { memo, useState, useCallback, useEffect } from 'react';

const query = (path, options) => fetch(path, {
		...options,
		headers: {
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	})
	.then(r => (r.headers.get("Content-Type") == "application/json" ? r.json() : r.text())
		.then(data => ([r.ok, data]))
	)
	.then(([ok, data]) => {
		if (ok) {
			return data
		} else {
			throw new Error(data)
		}
	})

const ITEMS = {
	Get: () => query(`/items/`)
		.then(items => {
			for (const id in items) {
				items[id].Time = new Date(items[id].Time)
			}
			return items
		}),
	Post: (item) => query(`/items/`, {
		method: "POST",
		body: JSON.stringify(item)
	}),
	Patch: (i, item) => query(`/items/${i}`, {
		method: "PATCH",
		body: JSON.stringify(item)
	}),
	Delete: (i) => query(`/items/${i}`, {
		method: "DELETE",
	}),
}

function Item({ index, item, patch, remove }) {
	const [pending, setPending] = useState(false)
	return (
		<tr
			style={{
				textDecorationLine: item.Done ? 'line-through' : 'none'
			}}
		>
			<td>
				{index}
			</td>
			<td>
				{item.Description}
			</td>
			<td>
				<input
					type="datetime-local"
					disabled={true}
					value={item.Time.toISOString().replace(/[zZ]$/, "")}
				/>
			</td>
			<td>
				<input
					type="checkbox"
					disabled={pending}
					checked={item.Done}
					onChange={(e) => {
						const alt = {
							...item,
							Done: e.target.checked,
						}
						setPending(true)
						patch(alt)
							.then(() => setPending(false))
					}}
				/>
			</td>
			<td>
				<button
					disabled={pending}
					onClick={() => {
						setPending(true)
						remove()
							.catch(e => console.log(e))
							.finally(() => setPending(false))
					}}
				>
					X
				</button>
			</td>
		</tr>
	)
}

function Draft({items, setItems}) {
	const [pending, setPending] = useState(false)
	const [draft, setDraft] = useState({
		Description: "",
		Done: false,
		Time: new Date(Date.now()),
	})

	const submit = useCallback(
		(e) => {
			setPending(true)
			ITEMS.Post(draft)
				.catch(e => console.log(e))
				.then(() => ITEMS.Get())
				.then(items => setItems(items))
				.catch(e => console.log(e))
				.finally(() => setPending(false))
		},
		[ draft, setPending, setItems ]
	)

	return (
		<tr>
			<td>
				-
			</td>
			<td>
				<input
					disabled={pending}
					onKeyDown={(e) => {
						if ((e.key) === 'Enter') {
							submit(e)
						}
					}}
					onChange={(e) => {
						setDraft({
							...draft,
							Description: e.target.value,
						})
					}}
				/>
			</td>
			<td>
				<input
					type="datetime-local"
					value={draft.Time.toISOString().replace(/[zZ]$/, "")}
					disabled={pending}
					onChange={(e) => {
						const time = new Date(e.target.value+"Z")
						if (time === "Invalid Date") {
							return
						}
						setDraft({
							...draft,
							Time: time,
						})
					}}
				/>
			</td>
			<td>
			</td>
			<td>
				<button
					disabled={pending || !draft.Description }
					onClick={submit}
				>+</button>
			</td>
		</tr>
	)
}

function Body({items, setItems, pending, setPending}) {
	const size = Object.keys(items).length
	const body = new Array(size)

	const patch = useCallback(
		(id, item) => ITEMS.Patch(id, item)
			.then(ITEMS.Get)
			.then(setItems),
		[setItems]
	)

	const remove = useCallback(
		(id) => ITEMS.Delete(id)
			.then(ITEMS.Get)
			.then(setItems),
		[setItems]
	)

	let i = 0;
	for (const id in items) {
		body[size-(i+1)] = (
			<Item
				index={i}
				key={id}
				item={items[id]}
				patch={patch.bind(this, id)}
				remove={remove.bind(this, id)}
			/>
		)
		i++
	}
	return body
}

function Todo() {
	const [items, setItems] = useState({})

	useEffect(() => {
		ITEMS.Get().then(setItems)
	}, [setItems])

	return (
		<React.Fragment>
			<h1>TODO</h1>
			<table>
				<thead>
					<Draft
						items={items}
						setItems={setItems}
					/>
				</thead>
				<tbody>
					<Body
						items={items}
						setItems={setItems}
					/>
				</tbody>
				<tfoot>
					<tr>
						<th>index</th>
						<th>description</th>
						<th>date</th>
						<th>done</th>
						<th>remove</th>
					</tr>
				</tfoot>
			</table>
		</React.Fragment>
	);
}

export default Todo;
