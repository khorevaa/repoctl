# Maintainer: Ben Morgan <neembi@gmail.com>
# vim: set ts=2 sw=2:
pkgname=repoctl
pkgver=0.12
pkgrel=1
pkgdesc="A supplement to repo-add and repo-remove which simplifies managing local repositories"
arch=('i686' 'x86_64')
url="https://github.com/cassava/repoctl"
license=('MIT')
depends=('pacman')
makedepends=('go')
source=(https://github.com/downloads/cassava/$pkgname/$pkgname-$pkgver.tar.gz)

build() {
  # Get and build the builder.
  mkdir ${srcdir}/go
  GOPATH=${srcdir}/go go get github.com/constabulary/gb/...

  cd $srcdir/$pkgname-$pkgver
  ${srcdir}/go/bin/gb build
}

package() {
  cd $srcdir/$pkgname-$pkgver

  # Install repo program
  mkdir -p $pkgdir/usr/bin
  install -m755 repoctl $pkgdir/usr/bin/

  # Install other documentation
  install -m644 TODO README.md NEWS $pkgdir/usr/share/doc/repo-keep/

  # Install completion files
  mkdir -p $pkgdir/usr/share/zsh/site-functions/
  install -m644 contrib/zsh_completion $pkgdir/usr/share/zsh/site-functions/_repoctl
}
