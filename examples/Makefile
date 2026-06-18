EXAMPLES := \
  query \
  update \
  prompt \
  merge \
  color \

ROOT := ..

build: $(EXAMPLES:%=build-%)

run: $(EXAMPLES:%=run-%)

clean: $(EXAMPLES:%=clean-%)

$(EXAMPLES:%=build-%):
	@$(MAKE) --no-print-directory -C $(ROOT) build-example EXAMPLE="$(@:build-%=%)"

$(EXAMPLES:%=run-%):
	@echo === Testing $@ ===
	@$(MAKE) --no-print-directory -C $(ROOT) run-example EXAMPLE="$(@:run-%=%)"
	@echo

$(EXAMPLES:%=clean-%):
	@$(MAKE) --no-print-directory -C $(@:clean-%=%) clean
