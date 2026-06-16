/*
 * libc_symtab.h — Compile-time libc symbol table for Linux BOF loading
 *
 * Go's internal linker does not export libc symbols in the dynamic symbol
 * table, so dlsym(NULL, "puts") returns NULL even though libc is loaded.
 * GCC resolves these pointers at compile time via CGo, so they are always
 * valid.  This is the Linux equivalent of bofdefs.h / DFR on Windows.
 *
 * To add support for a new libc function:
 *   1. Add the required #include if not already present.
 *   2. Add a {"name", (void*)name} entry in the appropriate section.
 *   3. If the function may be a macro (ctype.h, etc.), add a _wrap_ shim.
 */
#ifndef _LIBC_SYMTAB_H_
#define _LIBC_SYMTAB_H_

#if defined(__linux__)

/* ── Standard C ─────────────────────────────────────────────── */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdarg.h>
#include <ctype.h>
#include <errno.h>
#include <math.h>

/* ── POSIX / Linux ──────────────────────────────────────────── */
#include <unistd.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/statvfs.h>
#include <sys/sysinfo.h>
#include <sys/mman.h>
#include <sys/ioctl.h>
#include <sys/wait.h>
#include <sys/time.h>
#include <sys/utsname.h>
#include <dirent.h>
#include <pwd.h>
#include <grp.h>
#include <signal.h>
#include <time.h>
#include <utime.h>
#include <fnmatch.h>
#include <libgen.h>
#include <mntent.h>

/* ── Network ────────────────────────────────────────────────── */
#include <sys/socket.h>
#include <netdb.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <ifaddrs.h>
#include <net/if.h>

/* ── ctype wrappers (may be macros in glibc) ────────────────── */
static int _wrap_isdigit(int c) { return isdigit(c); }
static int _wrap_isprint(int c) { return isprint(c); }
static int _wrap_tolower(int c) { return tolower(c); }
static int _wrap_toupper(int c) { return toupper(c); }
static int _wrap_isalpha(int c) { return isalpha(c); }
static int _wrap_isalnum(int c) { return isalnum(c); }
static int _wrap_isspace(int c) { return isspace(c); }

/* ── Symbol table ───────────────────────────────────────────── */

typedef struct { const char* name; void* addr; } libc_sym_entry_t;

static libc_sym_entry_t _libc_symbols[] = {

    /* stdio */
    {"puts",      (void*)puts},
    {"printf",    (void*)printf},
    {"fprintf",   (void*)fprintf},
    {"sprintf",   (void*)sprintf},
    {"snprintf",  (void*)snprintf},
    {"vprintf",   (void*)vprintf},
    {"vfprintf",  (void*)vfprintf},
    {"vsnprintf", (void*)vsnprintf},
    {"fopen",     (void*)fopen},
    {"fclose",    (void*)fclose},
    {"fread",     (void*)fread},
    {"fwrite",    (void*)fwrite},
    {"fgets",     (void*)fgets},
    {"fputs",     (void*)fputs},
    {"fgetc",     (void*)fgetc},
    {"fseek",     (void*)fseek},
    {"ftell",     (void*)ftell},
    {"fflush",    (void*)fflush},
    {"rewind",    (void*)rewind},
    {"fdopen",    (void*)fdopen},
    {"fileno",    (void*)fileno},
    {"fscanf",    (void*)fscanf},
    {"sscanf",    (void*)sscanf},
    {"popen",     (void*)popen},
    {"pclose",    (void*)pclose},
    {"perror",    (void*)perror},

    /* stdlib */
    {"malloc",    (void*)malloc},
    {"calloc",    (void*)calloc},
    {"realloc",   (void*)realloc},
    {"free",      (void*)free},
    {"atoi",      (void*)atoi},
    {"atol",      (void*)atol},
    {"atof",      (void*)atof},
    {"strtol",    (void*)strtol},
    {"strtoul",   (void*)strtoul},
    {"strtoll",   (void*)strtoll},
    {"strtoull",  (void*)strtoull},
    {"qsort",     (void*)qsort},
    {"bsearch",   (void*)bsearch},
    {"abs",       (void*)abs},
    {"exit",      (void*)exit},
    {"_exit",     (void*)_exit},
    {"getenv",    (void*)getenv},
    {"setenv",    (void*)setenv},
    {"unsetenv",  (void*)unsetenv},
    {"system",    (void*)system},

    /* string */
    {"memcpy",     (void*)memcpy},
    {"memset",     (void*)memset},
    {"memmove",    (void*)memmove},
    {"memcmp",     (void*)memcmp},
    {"memchr",     (void*)memchr},
    {"strlen",     (void*)strlen},
    {"strcmp",     (void*)strcmp},
    {"strncmp",    (void*)strncmp},
    {"strcpy",     (void*)strcpy},
    {"strncpy",    (void*)strncpy},
    {"strcat",     (void*)strcat},
    {"strncat",    (void*)strncat},
    {"strdup",     (void*)strdup},
    {"strstr",     (void*)strstr},
    {"strchr",     (void*)strchr},
    {"strrchr",    (void*)strrchr},
    {"strerror",   (void*)strerror},
    {"strtok",     (void*)strtok},
    {"strtok_r",   (void*)strtok_r},
    {"strcasecmp", (void*)strcasecmp},
    {"strcspn",    (void*)strcspn},
    {"strspn",     (void*)strspn},
    {"basename",   (void*)basename},

    /* ctype (wrapped) */
    {"isdigit",  (void*)_wrap_isdigit},
    {"isprint",  (void*)_wrap_isprint},
    {"tolower",  (void*)_wrap_tolower},
    {"toupper",  (void*)_wrap_toupper},
    {"isalpha",  (void*)_wrap_isalpha},
    {"isalnum",  (void*)_wrap_isalnum},
    {"isspace",  (void*)_wrap_isspace},

    /* unistd / posix */
    {"getuid",      (void*)getuid},
    {"geteuid",     (void*)geteuid},
    {"getgid",      (void*)getgid},
    {"getegid",     (void*)getegid},
    {"getpid",      (void*)getpid},
    {"getppid",     (void*)getppid},
    {"getcwd",      (void*)getcwd},
    {"chdir",       (void*)chdir},
    {"access",      (void*)access},
    {"unlink",      (void*)unlink},
    {"rmdir",       (void*)rmdir},
    {"mkdir",       (void*)mkdir},
    {"rename",      (void*)rename},
    {"link",        (void*)link},
    {"symlink",     (void*)symlink},
    {"readlink",    (void*)readlink},
    {"read",        (void*)read},
    {"write",       (void*)write},
    {"open",        (void*)open},
    {"close",       (void*)close},
    {"lseek",       (void*)lseek},
    {"dup",         (void*)dup},
    {"dup2",        (void*)dup2},
    {"pipe",        (void*)pipe},
    {"sleep",       (void*)sleep},
    {"usleep",      (void*)usleep},
    {"gethostname", (void*)gethostname},
    {"sysconf",     (void*)sysconf},

    /* stat */
    {"stat",    (void*)stat},
    {"lstat",   (void*)lstat},
    {"fstat",   (void*)fstat},
    {"chmod",   (void*)chmod},
    {"fchmod",  (void*)fchmod},
    {"chown",   (void*)chown},
    {"fchown",  (void*)fchown},
    {"statvfs", (void*)statvfs},

    /* dirent */
    {"opendir",  (void*)opendir},
    {"readdir",  (void*)readdir},
    {"closedir", (void*)closedir},

    /* pwd / grp */
    {"getpwuid", (void*)getpwuid},
    {"getpwnam", (void*)getpwnam},
    {"getgrgid", (void*)getgrgid},
    {"getgrnam", (void*)getgrnam},

    /* process */
    {"fork",    (void*)fork},
    {"execve",  (void*)execve},
    {"execvp",  (void*)execvp},
    {"waitpid", (void*)waitpid},
    {"kill",    (void*)kill},
    {"signal",  (void*)signal},

    /* socket */
    {"socket",      (void*)socket},
    {"connect",     (void*)connect},
    {"bind",        (void*)bind},
    {"listen",      (void*)listen},
    {"accept",      (void*)accept},
    {"send",        (void*)send},
    {"recv",        (void*)recv},
    {"sendto",      (void*)sendto},
    {"recvfrom",    (void*)recvfrom},
    {"setsockopt",  (void*)setsockopt},
    {"getsockopt",  (void*)getsockopt},
    {"getsockname", (void*)getsockname},
    {"getpeername", (void*)getpeername},
    {"inet_ntoa",   (void*)inet_ntoa},
    {"inet_addr",   (void*)inet_addr},
    {"inet_ntop",   (void*)inet_ntop},
    {"inet_pton",   (void*)inet_pton},
    {"htons",       (void*)htons},
    {"ntohs",       (void*)ntohs},
    {"htonl",       (void*)htonl},
    {"ntohl",       (void*)ntohl},

    /* dns / netdb */
    {"getaddrinfo",  (void*)getaddrinfo},
    {"freeaddrinfo", (void*)freeaddrinfo},
    {"gai_strerror", (void*)gai_strerror},
    {"getnameinfo",  (void*)getnameinfo},

    /* ifaddrs */
    {"getifaddrs",  (void*)getifaddrs},
    {"freeifaddrs", (void*)freeifaddrs},

    /* time */
    {"time",          (void*)time},
    {"localtime",     (void*)localtime},
    {"gmtime",        (void*)gmtime},
    {"strftime",      (void*)strftime},
    {"clock_gettime", (void*)clock_gettime},
    {"nanosleep",     (void*)nanosleep},
    {"gettimeofday",  (void*)gettimeofday},
    {"utime",         (void*)utime},

    /* mntent */
    {"setmntent", (void*)setmntent},
    {"getmntent", (void*)getmntent},
    {"endmntent", (void*)endmntent},

    /* sys */
    {"mmap",     (void*)mmap},
    {"munmap",   (void*)munmap},
    {"mprotect", (void*)mprotect},
    {"ioctl",    (void*)ioctl},
    {"uname",    (void*)uname},
    {"sysinfo",  (void*)sysinfo},
    {"fnmatch",  (void*)fnmatch},

    /* sentinel */
    {NULL, NULL}
};

#define LIBC_SYM_COUNT (sizeof(_libc_symbols) / sizeof(_libc_symbols[0]) - 1)

#endif /* __linux__ */
#endif /* _LIBC_SYMTAB_H_ */
