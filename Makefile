default: clean package

clean:
	rm -f example-cc.tar.gz

package:
	tar -zcvf example-cc.tar.gz -C . *
