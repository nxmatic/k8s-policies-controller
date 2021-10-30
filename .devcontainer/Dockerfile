FROM qmcgaw/godevcontainer:debian as builder

RUN --mount=from=devcontainer/library-scripts,ro,target=.library-scripts \
    .library-scripts/update-cache-debian.sh && \
    .library-scripts/utils-debian.sh && \
    .library-scripts/direnv-debian.sh && \
    .library-scripts/clean-cache-debian.sh && \
    .library-scripts/dockerize.sh && \
    .library-scripts/docker-compose.sh 

COPY --from=devcontainer/library-scripts /devcontainer-postcreate.sh /devcontainer-postcreate.sh
COPY --from=devcontainer/library-scripts /git-clone.sh /etc/docker-postcreate.d/01-git-clone

COPY --from=devcontainer/library-scripts /docker-entrypoint.sh /

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD [ "sleep", "infinity"]