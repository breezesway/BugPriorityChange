# BugPriorityChange
This repository stores the raw data, code and calculation results for *Bug Priority Change: An Empirical Study on Apache Projects*.  
Specifically, the raw data includes all issue data obtained from JIRA and commit record files parsed from the code repository. 
The following will explain the replication package we provide.
## The Raw Data  
### JIRA Issue
We obtained all the issue data from JIRA and stored it in MySQL. Now we will share the issue data used in this paper. We export all the issue data into eight SQL files (one for each table), compress them into zip files, and upload them to Zenodo. The issue data can be accessed at [![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.8237407.svg)](https://doi.org/10.5281/zenodo.8237407).
> Note: The compressed package size is about 1.9G. If you want to use this data, after downloading and extracting the file, please create a database in MySQL and execute these SQL files.
### ProjectCommit
The directory named *ProjectCommit* lists the commit record files for the 32 projects we studied.
> Note: Some projects (such as *Cordova*) correspond to multiple commit record files, as these projects have multiple code repositories, each corresponding to a commit record file.
## Calculation Code
The directory named *GoPriorityChanged* is the relevant code for our calculation data, which is a program written in Golang. If you want to run it, please install Go 1.18.0 or higher version first.

## Calculation Results
In the directory named *RQresults*, we provide the calculation results of all RQs in the paper. Each RQ calculation result corresponds to a CSV file, and you can easily know which CSV file corresponds to which RQ through the file name.

## The init list of projects
For the case selection in the paper, we used all Apache projects as the initial list of projects for screening, and ultimately obtained the 32 projects in the paper. We have listed the initial projects in **the initial list of projects file**.

---
If you have any questions, please contact caigz1999@foxmail.com.
