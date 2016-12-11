# go-maven
The release process is pretty straightforward:
1. Ready for release? Make sure you've checked everything in.
2. Decide what version you want to release at.
3. Set the release version in the POM.
4. Package with Tests.
5. Deploy to the "release" repository.
6. Check in the POM changes.
7. Tag the release change in the SCM.
8. Decide what the new development version is.
9. Set the development version in the POM.
10. Check in the POM changes.

The maven-release-plugin is supposed to do exactly that, and a little bit more (like build 
and deploy the new development version). In my experience, though, it's too complicated, 
and doesn't actually work with a local Nexus repository (Step #5 doesn't happen). So I 
wrote this script in perl to simplify the use of the maven interface, and to actually do 
the steps outlined above.

This is probably NOT the "approved" maven way, or even the "approved" Perl way. A complete 
discussion of why Maven doesn't seem to be the best tool for this job will have to go 
elsewhere.
 
## Usage
 
     > go (default, equivalent to mvn test)
     > go [clean] [notest] [build|package|install|deploy]
     > go release
     
