# OCFL Validation

The `validation` package provides structures and functions for handling OCFL validation errors and warnings. It maps OCFL specification error codes to descriptive error objects and provides a mechanism to collect these errors during the validation process using Go's `context.Context`.

## Key Components

### ErrorCode
A string type representing an OCFL validation error or warning code (e.g., `E001`, `W001`).

### Error
A struct that contains detailed information about a validation issue:
- `Code`: The `ErrorCode`.
- `Description`: The official description from the OCFL specification.
- `Ref`: A URL to the relevant section in the OCFL specification.
- `Description2`: Additional dynamic information about the specific occurrence of the error.
- `Context`: The architectural context where the error occurred (e.g., "storage root", "object root").
- `Version`: The OCFL version the error refers to.

### Status
A structure that collects all validation errors and warnings encountered during a validation run.

## Usage

### Initializing Validation Context
To start a validation process, create a new context that carries a `Status` object:

```go
ctx := validation.NewContextValidation(context.Background())
```

### Adding Errors and Warnings
Errors and warnings are added to the context using `AddValidationErrors` and `AddValidationWarnings`. You can retrieve the base error definition using `GetValidationError` and then append specific details.

```go
vErr := validation.GetValidationError(version.OCFLVersion("1.1"), validation.E001)
vErr = vErr.AppendContext("object root").AppendDescription("found unexpected file: %s", filename)
validation.AddValidationErrors(ctx, vErr)
```

### Retrieving Validation Results
After the validation process is complete, retrieve the `Status` from the context:

```go
status, err := validation.GetValidationStatus(ctx)
if err != nil {
    // handle error (context didn't have validation status)
}

// Optionally compact to remove duplicates and sort
status.Compact()

for _, vErr := range status.Errors {
    fmt.Println(vErr.Error())
}
```

## Error Definitions

The package includes comprehensive mappings for OCFL validation codes. These mappings include the official description and a reference link to the specification.

### OCFL Version 1.0
- **Validation Codes:** [OCFL v1.0 Validation Codes](https://ocfl.io/1.0/spec/validation-codes.html)
- **Source:** `validationerror1_0.go`

### OCFL Version 1.1
- **Validation Codes:** [OCFL v1.1 Validation Codes](https://ocfl.io/1.1/spec/validation-codes.html)
- **Source:** `validationerror1_1.go`

These definitions are automatically initialized into `errorDetails` for integration with broader error handling systems.
