# patchenv

### Set values in the running process's environment using an external command

`patchenv` is a simple Go library with one public function that updates the
running process's environment with the variables it parses from the output of
a command you specify.

This means you can inject a dynamic set of environment variables into the
process without changing the environment of the process that starts the
process or the command line arguments used to start it. Integrated development
environments (IDEs) often make it inconvenient or difficult to inject dynamic
values in those configuration elements, so patchenv can help there.

### Usage

Add a requirement for `patchenv` to your `go.mod` file:

    require (
        ...
        github.com/arpio/patchenv v1.0.0
    )

Then, call `patchenv.Patch()` somewhere early in your program:

    package main
   
    import (
        "fmt"
        "github.com/arpio/patchenv"
        "os"
    )
   
    func main() {
        patchenv.Patch()
        fmt.Println(os.Environ())
    }

When `patchenv.Patch()` gets called, if the `PATCH_ENV_COMMAND` environment
variable is set, its value is executed as a shell command and the output of
that command is used to update the environment. Before you run your program,
set `PATCH_ENV_COMMAND` to a command that emits `KEY=value` lines for each
environment variable you want.

To set `FOO` to `bar`:

    PATCH_ENV_COMMAND="echo FOO=bar" ./myprogram
    [PATH=/usr/local/sbin:... HOME=/home/user FOO=bar]

So basically, set `PATCH_ENV_COMMAND` when you want `patchenv` to patch things
up for you, and don’t set it when you don’t.

Your command’s output should contain one environment variable per line, in the
format KEY=value:

    FOO=bar
    AWS_SESSION_TOKEN=FwoGZXIvY...
    HINT=values can have spaces and "special chars", but not newlines

#### Example: IntelliJ IDEA debugging with aws-vault

You're developing a program that uses the
[AWS SDK for Go](https://aws.amazon.com/sdk-for-go/) to access Amazon Web
Services (AWS). Your organization prohibits storing unencrypted access keys on
disk, so you use [aws-vault](https://github.com/99designs/aws-vault) to manage
them securely. This works great when you're running your program from the
command line, but there isn't an easy way to get your IDE to feed the output
of `aws-vault` into the Go process it starts.

Here's how you can use `patchenv` with an IDE like IntelliJ to inject
`aws-vault`'s output into the Go program you're debugging:

1. Add a requirement for `patchenv` to your program's `go.mod` file
2. Call `patchenv.Patch()` early in your program's `main` function or an
   initialization function
4. Edit your IntelliJ debug configuration and set the `PATCH_ENV_COMMAND`
   environment variable:

       PATCH_ENV_COMMAND=aws-vault exec my-profile -- sh -c "env | grep ^AWS_"

   Adjust the `aws-vault` command line as needed for your profile, session
   duration, etc. The important part is that we make `aws-vault` execute a
   shell process that pipes all its environment variables through `grep` so we
   select only the AWS credential variables.

Now run the debugger. You can always step into `patchenv.Patch()` if you want
to see how it works or diagnose an issue with it.

### Limitations

If `aws-vault` doesn't already have valid credentials when you start
debugging, it may need to read things like your MFA token from standard input.
This will fail since `patch_env` doesn't feed any input to its
`PATCH_ENV_COMMAND`.

As a work-around, open a new terminal and run `aws-vault exec` for the profile
you use for debugging, enter the credentials there, and then re-launch the
debugger.  `aws-vault` stores its session tokens in your system's keystore, so
they'll be available to other instances of `aws-vault` until they expire.

### Documentation

Package documentation is available at
[godoc](https://godoc.org/github.com/arpio/patchenv).

### See Also

For Python programs, see [patch_env](https://pypi.org/project/patch-env/),
which works similarly.
