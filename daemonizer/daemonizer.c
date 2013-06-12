/* s. rannou <mxs@sbrk.org> */

#define _BSD_SOURCE
#define _XOPEN_SOURCE

#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <pwd.h>

/*
 * Since GO isn't able to daemonize for now (because of potential
 * threading issues), we need to daemonize from another program.
 * start-stop-daemon is a good answer, unfortunately it's not
 * available everywhere and implementations are different from one
 * distribution to another.
 */

void usage(char *progname);

void daemonize()
{
    int pid;
    int i;

    /* let the parent die so we become orphan */
    pid = fork();
    if (pid < 0) {
        fprintf(stderr, "Unable to fork: %s\n", strerror(errno));
        exit(EXIT_FAILURE);
    }
    if (pid > 0) {
        exit(EXIT_SUCCESS);
    }

    /* get a new process group */
    if (setsid() == -1) {
        fprintf(stderr, "Unable to get a new session id: %s\n", strerror(errno));
        exit(EXIT_FAILURE);
    }

    /* get rid of open descriptors and redirect standard ones to dev null */
    for (i = getdtablesize(); i >= 0; --i) {
        close(i);
    }
    i = open("/dev/null", O_RDWR);
    if (i != -1) {
        dup(i);
        dup(i);
    }
}

uid_t get_uid_for_user(char *user)
{
    struct passwd *pw_user;

    if (user == NULL) {
        return getuid();
    }

    pw_user = getpwnam(user);
    if (!pw_user) {
        fprintf(stderr, "Unable to find the user %s: %s\n", user, strerror(errno));
        exit(EXIT_FAILURE);
        return -1;
    }

    return pw_user->pw_uid;
}

void mutate(uid_t uid)
{
    /* do not mutate in what we already are */
    if (uid != getuid()) {
        if (setuid(uid) == -1) {
            fprintf(stderr, "Unable to setuid to user %d: %s\n", uid, strerror(errno));
            exit(EXIT_FAILURE);
        }
    }
}

void jail(char *dest)
{
    if (!dest) {
        return;
    }

    /* at this point, we have already chdired to dest */
    if (chroot(".") == -1) {
        fprintf(stderr, "CANT CHROOT?\n");
        fprintf(stderr, "Unable to chroot to %s: %s\n", dest, strerror(errno));
        exit(EXIT_FAILURE);
    }
}

void run(char *argv[], char *pidfile)
{
    char pidstr[32], *pidpos;
    int fd, to_write, ret;

    fd = open(pidfile, O_CREAT | O_RDWR, 0640);
    if (fd == -1) {
        fprintf(stderr, "Unable to open pidfile %s: %s\n", pidfile, strerror(errno));
        exit(EXIT_FAILURE);
    }
    if (lockf(fd, F_TLOCK, 0) == -1) {
        fprintf(stderr, "Unable to lock pidfile %s: %s\n", pidfile, strerror(errno));
        exit(EXIT_FAILURE);
    }

    /* write pid to pidfile */
    to_write = snprintf(pidstr, sizeof(pidstr), "%d\n", getpid());
    pidpos = pidstr;
    while (to_write > 0) {
        ret = write(fd, pidpos, to_write);
        if (ret == -1) {
            fprintf(stderr, "Unable to write pid to pidfile %s: %s\n", pidfile, strerror(errno));
            exit(EXIT_FAILURE);
        }
        pidpos += ret;
        to_write -= ret;
    }

    if (execve(argv[0], argv, NULL) == -1) {
        unlink(pidfile);
        fprintf(stderr, "Unable to exec command %s: %s\n", argv[0], strerror(errno));
        exit(EXIT_FAILURE);
    }
}

int main(int argc, char *argv[])
{
    char *user = NULL;
    char *dir = NULL;
    char *pidfile = NULL;
    char *progname = argv[0];
    int has_jail_opt = 0;
    int has_debug_opt = 0;
    int opt;
    uid_t uid;

    while ((opt = getopt(argc, argv, "djp:u:c:")) != -1) {
        switch (opt) {
        case 'u':
            user = optarg;
            break;

        case 'd':
            has_debug_opt = 1;
            break;

        case 'p':
            pidfile = optarg;
            break;

        case 'c':
            dir = optarg;
            break;

        case 'j':
            has_jail_opt = 1;
            break;

        default:
            usage(progname);
            /* NEVER REACHED */
        }
    }

    if (!pidfile || (has_jail_opt && !dir)) {
        usage(progname);
    }

    argc -= optind;
    argv += optind;

    if (argc <= 0) {
        usage(progname);
        /* NEVER REACHED */
    }

    if (dir) {
        if (chdir(dir) == -1) {
            fprintf(stderr, "Unable to chdir to %s: %s\n", dir, strerror(errno));
            exit(EXIT_FAILURE);
        }
    }

    uid = get_uid_for_user(user);

    if (has_jail_opt) {
        jail(dir);
    }

    mutate(uid);

    if (!has_debug_opt) {
        daemonize();
    }

    run(argv, pidfile);

    exit(EXIT_SUCCESS);
}

void usage(char *progname)
{
    fprintf(stderr,
            "Usage : %s -p PIDFILE [-d] [-j] [-u USER] [-c CHDIR] -- COMMAND [ARGS]\n"
            "\n"
            "-p\t\tuse PIDFILE (relative to CHDIR) as a lock file in which the PID of COMMAND is written\n"
            "-d\t\tdebug mode, do not daemonize to print errors\n"
            "-j\t\tjail the command with chroot, has *no* effect if you aren't root\n"
            "-u\t\texecute COMMAND as USER\n"
            "-c\t\tchange the working directory of COMMAND to CHDIR\n",
            progname);
    exit(EXIT_FAILURE);
}
