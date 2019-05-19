## JSON Schema + Go + APIs

After working on a few Go HTTP Servers i've noticed 
a few different patterns for API validations. I'm going to attempt to show off _yet another_ 
way this can be done and hopefully convince you to love [JSON Schema](https://json-schema.org/) as much as I do.

### What is JSON Schema?
If you haven't heard of JSON Schema before, it's an attempt to apply structure to otherwise unstructured JSON.
As developers, sometimes "unstructured" JSON is incredibly convenient (looking at you [postgres](https://www.postgresql.org/docs/9.4/datatype-json.html)), 
but at some point in the lifecycle of that data, you will need to perform some validation on it. Whether that is checking if a certain field is present, the type of that field, if it's an integer between 1 and 10 exclusive, or something else.
This is where JSON Schema comes in.

### What can I do with JSON Schema?
Just about any data validations on each individual field. For example, you can validate:
* a field is an integer
* that integer is in the range [1-10)
* that field is present because it is _required_
* a field called "email" matches an email validation _regex_
* a field is a valid enum type

and much more (seriously, check all out all the things you can validate [here](http://json-schema.org/latest/json-schema-validation.html#rfc.section.6)).

### How can I have some?
Now that i've convinced you it's worth learning more about, let's dive into how we can add JSON Schema to a Go project.

One great feature of JSON Schema is that you can add validation to an existing project that using some implicit (or explicit) validations with no changes required. 

I'll be using [gojsonschema](https://github.com/xeipuuv/gojsonschema) to help with manipulating and validating our schemas.

#### The Schema
The first thing you need to do in either a new project or an existing one is create your schema.

Let's say we're building a blogging website. The API we're building allows the frontend to `POST` blog posts via HTTP.

In our blog website, we accept 7 `POST` parameters (or, in JSON Schema, `properties`), and some of them have validations. Our schema looks like:

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "title": "post",
  "description": "a blog post",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 50,
      "pattern": "^[A-Z].*"
    },
    "date": {
      "type": "string"
    },
    "body": {
      "type": "string"
    },
    "views": {
      "type": "integer",
      "minimum": 1
    },
    "post_type": {
      "type": "string",
      "enum": ["cross-post", "original"]
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "required": ["title", "date", "body", "post_type"]
}
```

In our schema, we define 4 of these properties to be required, and we apply validations to all of them - some more strict than others.

#### Loading our Schema

Now that we have written out our schema, we need to load our schema so we can use it to verify the `POST` body on every request.

I've chosen to check my schema into my project as a go string literal like:

```go
package main
const schemaJSON = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "title": "post",
  "description": "a blog post",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 50,
      "pattern": "^[A-Z].*"
    },
    "date": {
      "type": "string"
    },
    "body": {
      "type": "string"
    },
    "views": {
      "type": "integer",
      "minimum": 1
    },
    "post_type": {
      "type": "string",
      "enum": ["cross-post", "original"]
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "required": ["title", "date", "body", "author_email", "post_type"]
}`
```
 
but there are other ways to do this as well (gojsonschema [supports a static file on disk, Web/HTTP references, or custom Go types](https://github.com/xeipuuv/gojsonschema#loaders)).

#### Decorate a Handler
I tend to decorate my `http.Handler`s with middleware so I can use do this schema validation in a single place.
I wrote an `http.Handler` that looks like this:

```go
func validate(schema *gojsonschema.Schema, next http.HandlerFunc) http.HandlerFuc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		requestJSON := gojsonschema.NewBytesLoader(body)
		result, err := schema.Validate(requestJSON)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !result.Valid() {
			if err := writeError(result.Errors(), w); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
```

I won't go into too much detail about the closure, but if you are interested in learning more about writing HTTP servers + go, I highly recommend [this](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831) blog post from Mat Ryer.

This is where all the validation comes into play - with my loaded `schema` from the step before, 
we can simple do a `schema.Validate(requestJSON)`, and pass in a `JSONLoader`. 
If the JSON body is invalid, we will see `result.Valid()` is `false` and we can check each
of the errors in `result.Errors()`.

#### Wire everything up

I have another handler to actually fulfil the request logic that looks like this:

```go
func process(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("valid request"))
}
```


In the `main` function, all we need is to decorate our `http.Handler` and begin our `http.Server`:

```go
handler := validate(schema, process)
http.ListenAndServe(":8000", handler)
```

#### Conclusion


