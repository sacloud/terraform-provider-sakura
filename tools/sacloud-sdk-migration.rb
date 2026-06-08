#! /usr/bin/env ruby
# -*- coding: utf-8 -*-

$dir = `git -C #{__dir__} rev-parse --show-toplevel`.chomp
IO. popen("git -C #$dir ls-files -z") \
  . each_line("\0")                   \
  . map { _1.chomp("\0") }            \
  . grep(/\.go\z/)                    \
  . each                              \
do |name|
  path = File.join($dir, name)
  content = File.read(path, mode: 'r:utf-8')
  STDERR.puts "Processing #{name}..."

  content.gsub!(/^import \((.+?)\)/nm) do |import|
    lines = $1.dup
    lines.gsub!(%r|(\w+ )?"github.com/sacloud/([\w-]+)-api-go(?=[/"])|, '\\1"github.com/sacloud/sacloud-sdk-go/api/\\2')
    lines.gsub!(%r|(\w+ )?"github.com/sacloud/([\w-]+)-service-go(?=[/"])|, '\\1"github.com/sacloud/sacloud-sdk-go/service/\\2')
    lines.gsub!(%r|(\w+ )?"github.com/sacloud/saclient-go(?=[/"])|, '\\1"github.com/sacloud/sacloud-sdk-go/common/saclient')
    lines.gsub!(%r|(\w+ )?"github.com/sacloud/packages-go(?=[/"])|, '\\1"github.com/sacloud/sacloud-sdk-go/common/packages')
    
    next "import (#{lines})"
  end

  File.write(path, content, mode: 'w:utf-8')
end