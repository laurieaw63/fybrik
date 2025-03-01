# Logging

This page describes the information that your code should provide in all log entries it generates, and some tools fybrik provides to ensure consistency across components.

# Background
* Log entries should be written to stdout and stderr.
* Fybrik does not collect nor aggregate logs.  This may be done by external tools. (ex: logstash, fluentd, etc.)
* A globally unique identifier for each FybrikApplication instance is passed to all control plane and data plane components to be included in log entries.  This enables corrrelation of log entries across different logs and clusters for the specific instance, even if the name of the FybrikApplication is reused over time.

# Log Entry Contents
All fybrik components, whether control plane or data plane components, should write log entries to stdout and stderr in json format.
The contents of the log entries are detailed in fybrik.io/pkg/logging/logging.go.

The fybrik control plane uses [zerolog](github.com/rs/zerolog) for its golang components, and provides a library of fybrik specific helper functions to be used with it.
Examples of how to use zerolog: https://github.com/rs/zerolog/blob/master/log_example_test.go

TBD - fybrik logging helper functions for python and java.

# Log Entry Verbosity
Log levels should be used as follows:

- panic (zerolog.PanicLevel, 5) - Errors that prevent the component from operating correctly and handling requests
     Ex: fybrik control plane did not deploy correctly
	   Ex: Data plane component crashed and cannot handle requests
- fatal (zerolog.FatalLevel, 4) - Errors that prevent the component from successfully completing a particular task
	   Ex: fybrikapplication controller cannot generate a plotter
	   Ex: Arrow/Flight server used to read data cannot access data store
- error (zerolog.ErrorLevel, 3) - Errors that are not fatal nor panic, but that the user / request initiator is made aware of (typical production setting for stable solution)
	   Ex: Dataset requested in fybrikapplication.spec is not allowed to be used
 	   Ex: Query to Arrow/Flight server used to read data returns an error because of incorrect dataset ID
- warn (zerolog.WarnLevel, 2) - Errors not shared with the user / request initiator, typically from which the component recovers on its own
- info (zerolog.InfoLevel, 1) - High level health information that makes it clear the overall status, but without much detail (highest level used in production)
- debug (zerolog.DebugLevel, 0) - Additional information needed to help identify problems (typically used during testing)
- trace (zerolog.TraceLevel, -1) - For tracing step by step flow of control (typically used during development)

# Environment Variables
- LOGGING_VERBOSITY - should be set to one of the levels described in the previous section.  
- PRETTY_LOGGING - If true log entries are in human readable format.  If false, they are in json. Should only be true during individual development, since json is preferred to enable easy parsing by aggregator tools.

