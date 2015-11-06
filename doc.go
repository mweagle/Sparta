/*
The sparta package transforms a set of golang functions into a deployable package that includes:

	 	1. NodeJS proxy logic
	 	2. A golang binary
	 	3. Dynamically generated CloudFormation template that supports create/update & delete operations.
	 	4. If specified, CloudFormation custom resources to automatically configure S3/SNS push registration

See the Main() docs for more information.
*/
package sparta
