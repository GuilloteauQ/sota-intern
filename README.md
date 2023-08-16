# SotA Intern

a script aiming to generate a list of papers to read based on keywords


## Objective

```shell
sota-intern "self-scheduling" "openmp" ...
```

1. query the Hal and Arxiv apis for papers with those words in the title or abstracts
2. extract the references from those papers
3. for all the references available online, fetch the references
4. continue until no more or depth too deep
5. construct a ranking of most frequent papers cited 


## TODO

- [X] get papers from HAL from keywords
- [ ] get papers from HAL from title
- [ ] get papers from Arxiv from keywords
- [ ] get papers from Arxiv from title
- [ ] extract references names
- [ ] recursive calls of papers
- [ ] create ranking
