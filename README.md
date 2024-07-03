# Crestal Go Utils

This package provides a toolbox for crestal golang projects.  
In no particular case, we give preference to the [12-factor](https://12factor.net/) guidance.

## Logger
this package use standard slog package for logging, 
so if you set the global default logger anywhere, it will be used by this package.
```go
  slog.SetDefault(YOUR_LOGGER)
```
