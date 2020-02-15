# jg
`jg` is a [JSON](https://www.json.org/json-en.html) generator written in [Go](https://golang.org).  
It can be used to generate JSON documents of any complexity from human-friendly [YAML schema](#schema) specification.

## Why another JSON generator?
There are some other JSON generators:
* [GeneratedData](https://www.generatedata.com)
* [JSON Generator](https://www.json-generator.com)

Unfortunately, these generators are not flexible enough in strings generation.
They support only a fixed range of categories (e.g. *Name*, *City*, *Country*, etc...)
which are *internally* used to select appropriate wordlist to select strings from.
So there is no way to manually pass a dictionary to these generators.

`jg` supports passing external dictionaries with [files](#files) flag.

## Usage
```bash
jg [OPTIONS] SCHEMA
```

## Install


## Schema
Schema is defined in [YAML](https://yaml.org) format.  
There are two top-level fields:
* `root: node`
* `files: object`: a mapping with file names.  
  These names are **not** real paths in file system. They can be mapped to real files with [files](#files) CLI argument.
  ```yaml
  files:
    file1:
    file2:
    # ...
    fileN:
  ```

Each node (even `root`) must specify its type with `type` field:
```yaml
root:
  type: object
  fields:
    field1:
      type: int
```
Supported types:
* [bool](#bool)
* [int](#int)
* [float](#float)
* [string](#string)
* [object](#object)
* [array](#array)

Types [bool](#bool), [int](#int) and [float](#float) can be inlined.
In this case, the defaults are applied for each type correspondingly.
```yaml
boolInline: bool
boolExplicit:
  type: bool

integerInline: int
integerExplicit:
  type: int
  range: [0, 100]

floatInline: float
floatExplicit:
  type: float
  range: [0, 1]
```

### `bool`
A boolean value. It simply generates `true` or `false` in output JSON.

### `int`
An integer number. It can have only one of possible fields:
* `range: {int | [int, int]}` (default `[0, 10]`)  
  Range of posssible values (with maximum **included**). It can be one of the following types:
  * `int`: equivalent to `[0, int]`
    ```yaml
    range: 10 # [0, 10]
    ```
  * `[int, int]`: minimum and maximum correspondingly
    ```yaml
    range: [0, 10]
    ```
* `choices`
  List of `int`s
  ```yaml
  choices: [2, 3, 5, 7, 11, 13, 17, 19]
  ```

### `float`
An floating-point number. It can have only one of possible fields:
* `range: {int | [float, float]}` (default `[0, 1]`)  
  Interval of posssible values (with maximum **excluded**). It can be one of following types:
  * `float`: equivalent to `[0, float]`
    ```yaml
    range: 7.7
    ```
  * `[float, float]`: minimum and maximum correspondingly
    ```yaml
    range: [5.2, 11.3]
    ```
* `choices`: list of `float`s
  ```yaml
  choices: [3.14, 2.71, 4.20]
  ```

### `string`
A string value. It must specify one of the following fields:
* `from: string`: name of file to take strings from. This name should be listed in [files](#files) top-level field.
  ```yaml
  files:
    someFile:
  
  root:
    stringFromFile:
      type: string
      from: someFile
  ```
* `choices`: list of `string`s
  ```yaml
  choices:
    - choice 1
    - choice 2
    - choice 3
  ```


### `array`
An array object. It must specify its `elements`.
* `elements: node`
  Defines an element of array. It can be node of any type.
* `length: {uint | [uint, uint]}` (default: `10`)
  Length of array to generate. It can be one of the following types:
  * `uint`: exact length of array
    ```yaml
    length: 10
    ```
  * `[uint, uint]`: minimum and maximum of length correspondingly
    ```yaml
    length: [0, 10]
    ```

### `object`
An object. It must specify its `fields`:
* `fields: {obect}`: mapping of field names to nodes. Example:
  ```yaml
  type: object
  fields:
    a: int
    b: float
  ```

## Examples
Here is a simple schema:
```yaml
root:
  type: object
  fields:
    x: int
    x: int
```

See [examples](/examples) for more.
