# Use of `custom_func`, Specially `javascript`

`custom_func` is a transform that allows schema writer to alter, compose and transform existing
data from the input. Among [all `custom_func`](./customfuncs.md), `javascript` is the most important one to
understand and master.

A `custom_func` has 4 basic parts: `xpath`/`xpath_dynamic`, `name`, `args`, and `type`.

Like any other transforms, `custom_func` uses optional `xpath`/`xpath_dynamic` directive to move the current
IDR tree cursor. See [here](xpath.md#data-context-and-anchoring) for more details.

`name` is self-explanatory.

Optional `type` indicates a result type cast is needed. Valid types are `'string'`, `'int'`, `'float'`,
and `'boolean'`.

Now let's take an in-depth look at `custom_func.args`.

## `custom_func` Arguments

Each `custom_func` has different argument signature, 

## `custom_func` Composability

## `javascript`