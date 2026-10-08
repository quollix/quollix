# Quollix

All information about the project can be found on the [official website](https://quollix.org).

## Bootstrap

To begin developing, you need to install a few dependencies. On Ubuntu, this process is automated by a script. If you have a different operating system, however, you will need to install the dependencies manually. If you are an Ubuntu user, run the following command:

```bash
bash scripts/bootstrap.sh
```

Reboot the PC to complete the process.

## Getting started

On Ubuntu, add to `/etc/hosts` the entry:

```plain
127.0.0.1 quollix.localhost localhost sampleapp.localhost
```

Later, add similar entries for all apps you want to open with `<app>.localhost`. For example, `forgejo.localhost`.

Build and start the development environment:

```bash
cd src/ci-runner
go build
./ci-runner common artifacts # generate code for initial setup
./ci-runner deploy prod
```

Quollix should then be accessible in the browser at `https://quollix.localhost`. Sign in with username `admin` and password `password`.

To run the test suite from `src/ci-runner`:

```bash
./ci-runner test all
```

## Contributing

Please read the [Community](https://quollix.org/docs/project/community/) articles for more information on how to contribute to the project and interact with others.

## License

This project's source code is licensed under the [MIT License](LICENSE).

Brand assets covered by the [Quollix Brand Policy](BRAND_POLICY.md) are excluded from this license unless expressly stated otherwise.

Third-party assets are distributed under their own licenses. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
