# OCFL Validation

The `validation` package provides structures and functions for handling OCFL validation errors and warnings. It maps OCFL specification error codes to descriptive error objects.

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
A structure that collects all validation errors and warnings encountered during a validation run. Use `NewStatus()` to create one.

## Usage

### Collecting Errors
Validation errors are typically managed via the `OCFLLogger` in the `pkg/ocfllogger` package, which uses this package to create and store errors.

To manually create a validation error:
```go
vErr := validation.GetValidationError(version.OCFLVersion("1.1"), validation.E001)
vErr = vErr.AppendContext("object root").AppendDescription("found unexpected file: %s", filename)
```

To add it to a status collector:
```go
status := validation.NewStatus()
status.Add(vErr)
```

### Retrieving and Processing Results
After the validation process is complete, you can process the collected errors:

```go
// Remove duplicates and sort by context/code
status.Compact()

for _, vErr := range status.Errors {
    fmt.Println(vErr.Error())
}
```

## Error Definitions

The package includes comprehensive mappings for OCFL validation codes for different versions. These mappings include the official description and a reference link to the specification.

### OCFL Version 1.0
- **Source:** `validationerror1_0.go`
- **Specification:** [OCFL v1.0 Validation Codes](https://ocfl.io/1.0/spec/validation-codes.html)

### OCFL Version 1.1
- **Source:** `validationerror1_1.go`
- **Specification:** [OCFL v1.1 Validation Codes](https://ocfl.io/1.1/spec/validation-codes.html)
