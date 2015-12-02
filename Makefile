default: build

.PHONY: run edit

clean:
	rm -rf ./public

build: clean
	hugo

commit:
	git add --all . && git commit

commit-nomessage:
	git add --all . && git commit -m "Updated documentation"

pull:
	git pull origin docs

push:
	git push -f origin docs

edit: clean
	# Used for localhost editing
	hugo server --watch --verbose

publish: build commit push
	# Publish locally committed content to gh-pages
	# http://stevenclontz.com/blog/2014/05/08/git-subtree-push-for-deployment/
	# git push origin `git subtree split --prefix public docs`:gh-pages --force
	git subtree push --prefix public origin gh-pages
