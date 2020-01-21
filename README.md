# Pest-control
📬 Service for handling notification preferences

## API Documentation
The following APIs are protected by `heimdall`, so requests must have the
`Authorization` header set to the value `Bearer <token>`, where `<token>` is the
token generated by `heimdall`. `heimdall` would then forward the request with an
added `user_id` value that would be taken from the token.

### POST api/prefs
Creates a new set of preferences for a user.

#### Request body format
```
{
    "global": {
        "invitation": boolean (default: true),
        "text_entered": boolean (default: true),
        "text_modified": boolean (default: true),
        "tag": boolean (default: true),
        "role": boolean (default: true),
    },
    "conversation": [
        {
            "conversation_id": integer (default: 0),
            "text_entered": boolean (default: true),
            "text_modified": boolean (default: true),
            "tag": boolean (default: true),
            "role": boolean (default: true),
        }
    ]
}
```

All fields are optional. By default (i.e. if the request body is `{}`), all
`global` fields are `true` and there are no conversation notification
preferences.

#### Response body format
The body of a `200 OK` response will contain a representation of the created
resource. An example response body is shown below.
```
{
    "_id": string,
    "global": {
        "text_entered": true,
        "text_modified": true
    },
    "conversation": [
        {
            "conversation_id": 13,
            "text_entered": true
        }
    ]
}
```
A `409 Conflict` response will be returned if preferences already exist for the
user.

### POST api/prefs/conversations
Creates new conversation preferences for a user.

#### Request body format
```
{
    "conversation_id": integer (default: 0, required),
    "text_entered": boolean (default: true, optional),
    "text_modified": boolean (default: true, optional),
    "tag": boolean (default: true, optional),
    "role": boolean (default: true, optional),
}
```

By default (for example, if the request body is `{"conversation_id":2}`), all
boolean fields are `true`.

#### Response body format
The body of a `200 OK` response will contain a representation of the created
resource. An example response body is shown below.
```
{
    "conversation_id": 13,
    "text_entered": true
}
```
A `409 Conflict` response will be returned if preferences already exist for the
user.

### GET api/prefs
Retrieves global user preferences.

#### Response body format
The body of a `200 OK` response will contain a representation of the queried
resource. An example response body is shown below.
```
{
    "invitation": true,
    "text_entered": true,
    "text_modified": true
}
```
A `404 Not Found` response will be returned, if the user's preferences do not
exist, with a body that is a string indicating the error.

### GET api/prefs/conversations/{conversation_id}
Retrieves user preferences for a specific conversation.

#### Response body format
The body of a `200 OK` response will contain a representation of the queried
resource. An example response body is shown below.
```
{
    "text_entered": true,
    "text_modified": true
}
```
A `404 Not Found` response will be returned, if the user's preferences do not
exist, with a body that is a string indicating the error.

### DELETE api/prefs
Deletes user's preferences.

#### Response body format
A successful deletion will result in a `204 No Content` response with no body.
If the user's preferences do not exist, the response will have a status of `404
Not Found` and a body that is a string indicating the error.

### DELETE api/prefs/conversations/{conversation_id}
Deletes user's preferences for a specific conversation.

#### Response body format
A successful deletion will result in a `204 No Content` response with no body.
If the user's preferences for the conversation does not exist, the response will
have a status of `404 Not Found` and a body that is a string indicating the
error.

### PATCH api/prefs
Updates the global preferences of a user.

#### Request body format
```
{
    "invitation": boolean (default: true),
    "text_entered": boolean (default: true),
    "text_modified": boolean (default: true),
    "tag": boolean (default: true),
    "role": boolean (default: true),
}
```

All fields are optional. By default (i.e. if the request body is `{}`), all
fields are `true`.

#### Response body format
The body of a `200 OK` response will contain a representation of the updated
resource. An example response body is shown below.
```
{
    "invitation": true,
    "text_entered": true,
    "text_modified": true
}
```
A `404 Not Found` response will be returned, if the user's preferences do not
exist, with a body that is a string indicating the error.
