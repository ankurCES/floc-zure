# azfloci version

Print the azfloci version string.

## Synopsis
```
azfloci version
```

## Description
Prints the build version. Set at compile time via `-ldflags "-X .../internal/cli.Version=..."`. Shows `dev` for local builds without tags.

## Examples
```bash
azfloci version
# azfloci v0.1.0
```

## Related Commands
- `azfloci --help`
