
# Greenhouse

Integrated solidity environment.

## Tutorial

Start an empty folder and initialize it:

```
$ mkdir greenhouse-example
$ cd greenhouse-example
$ greenhouse init
Config file created
```

Create a simple smart contract:

```
$ mkdir contracts
$ cat <<EOT >> contracts/Simple.sol
pragma solidity >=0.8.0;

contract Simple {
}
EOT
```

Build the smart contract:

```
$ greenhouse build
```
