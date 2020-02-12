NAME = $(notdir $(PWD))

VERSION = $(shell printf "%s.%s" \
	$$(git rev-list --count HEAD) \
	$$(git rev-parse --short HEAD) \
)

version:
	@echo "$(VERSION)"

image:
	@echo :: building image $(NAME):$(VERSION)
	docker build -t $(NAME):$(VERSION) --build-arg version=$(VERSION) -f Dockerfile .

push:
	$(if $(REMOTE),,$(error REMOTE is not set))
	@echo :: pushing image $(NAME):$(VERSION)
	docker tag $(NAME):$(VERSION) $(REMOTE)/$(NAME):$(VERSION)
	docker push $(REMOTE)/$(NAME):$(VERSION)
	docker tag $(NAME):$(VERSION) $(REMOTE)/$(NAME):latest
	docker push $(REMOTE)/$(NAME):latest
