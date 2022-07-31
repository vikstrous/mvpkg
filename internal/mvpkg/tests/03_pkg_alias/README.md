# Package alias

This tests moves and renames a package that is imported under a package alias by another package.
The test ensures we don't rename identifiers that equal the package we're moving, but are alias.

We move ./source/target to ./destination/targetnew.
