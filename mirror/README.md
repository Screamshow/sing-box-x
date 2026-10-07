# Mirror publication

Install `sync.py` on the existing Forkop mirror and run:

```sh
python3 sync.py --tag 1.0.0
```

The publisher verifies GitHub SHA-256 digests and the upstream manifest before signing APK copies with the existing mirror key. It regenerates hashes after signing, preserves upstream hashes and installed-size metadata, and publishes an immutable release directory atomically. IPK and binary archive hashes are unchanged. Private keys stay on the mirror.

Preliminary releases update `candidate.json`. Stable releases update `latest.json`; no existing Forkop channel, package feed or installer default is changed. Router URLs in all published package records point to the mirror. Re-running verifies existing files and refuses changed upstream content under the same version.
