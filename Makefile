SPARTA_TEMP_DIR = ./.sparta
HUGO_BINARY := $(SPARTA_TEMP_DIR)/hugo
UNAME := $(shell uname)
HUGO_TARGZ_ARCHIVE_URL := ""
ifeq ($(UNAME), Linux)
	HUGO_TARGZ_ARCHIVE_URL="https://github.com/gohugoio/hugo/releases/download/v0.31.1/hugo_0.31.1_Linux-64bit.tar.gz"
endif
ifeq ($(UNAME), Darwin)
	HUGO_TARGZ_ARCHIVE_URL="https://github.com/gohugoio/hugo/releases/download/v0.31.1/hugo_0.31.1_macOS-64bit.tar.gz"
endif


default: build

.PHONY: install_hugo
install_hugo:
ifneq ("$(wildcard $(HUGO_BINARY))","")
	echo "Hugo already installed at: $(HUGO_BINARY)"
else
	mkdir -pv $(SPARTA_TEMP_DIR)
	curl -L -o $(SPARTA_TEMP_DIR)/hugo.tar.gz $(HUGO_TARGZ_ARCHIVE_URL)
	tar -xvf $(SPARTA_TEMP_DIR)/hugo.tar.gz -C $(SPARTA_TEMP_DIR)
	rm -rf $(SPARTA_TEMP_DIR)/hugo.tar.gz
endif
	$(SPARTA_TEMP_DIR)/hugo version

clean:
	rm -rf ./public

build: clean
	$(HUGO_BINARY)

test: install_hugo
	echo "Hugo installed"

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
	$(HUGO_BINARY) server --watch --verbose

publish: build commit push
	# Publish locally committed content to gh-pages
	# http://stevenclontz.com/blog/2014/05/08/git-subtree-push-for-deployment/
	# git push origin `git subtree split --prefix public docs`:gh-pages --force
	# If you run into "Local behind remote, can't fast forward":
	# git push origin `git subtree split --prefix public gh-pages`:gh-pages --force
	git subtree push --prefix public origin gh-pages
