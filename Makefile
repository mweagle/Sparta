default: build

.PHONY: run edit

clean:
	rm -rf ./public

build: clean
	hugo

test:
	rm -rfv ./tmp
	mkdir -pv ./tmp
	# Github replies with a 302 and location header response
	curl -L -o ./tmp/hugo.tar.gz https://github.com/spf13/hugo/releases/download/v0.15/hugo_0.15_linux_amd64.tar.gz
	tar -xvf ./tmp/hugo.tar.gz -C ./tmp
	./tmp/hugo_0.15_linux_amd64/hugo_0.15_linux_amd64
	rm -rfv ./tmp

reset: clean
		git reset --hard
		git clean -f -d

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
	# Windows: ./hugo.exe server --watch --verbose --renderToDisk
	hugo server --watch --verbose

publish: build commit push
	# Publish locally committed content to gh-pages
	# http://stevenclontz.com/blog/2014/05/08/git-subtree-push-for-deployment/
	# git push origin `git subtree split --prefix public docs`:gh-pages --force
	# If you run into "Local behind remote, can't fast forward":
	# git push origin `git subtree split --prefix public gh-pages`:gh-pages --force
	git subtree push --prefix public origin gh-pages
