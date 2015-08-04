ALL = zfssnap

all: $(ALL)

zfssnap: zfssnap.go
	env GOPATH=`pwd` go build $^

clean:
	-rm -f $(ALL) *~ core *.da *.gcov *.bb *.bbg gmon.out *.o
