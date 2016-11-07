# Configuration of CI

The project relies on [Concourse CI](https://concourse.ci) to run the tests and build the artifacts.

The pipeline contains template variables which are stored in [ansible vault](http://docs.ansible.com/ansible/playbooks_vault.html) format.

To set the pipeline, run this command:
```bash
fly --target <target-name> set-pipeline -p <pipeline-name> -c ci/pipeline.yml -l <(ansible-vault --vault-password-file=<path/to/vault/pass/file> decrypt ci/vars.yml --output=-)
```

If you need to change variable, ask project maintainers for a vault password file.
