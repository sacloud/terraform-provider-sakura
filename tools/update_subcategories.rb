require 'yaml'

tool_dir = File.dirname(File.expand_path(__FILE__))
docs_dir = File.expand_path(File.join(tool_dir, '..', 'docs'))
subcategory_file = File.join(tool_dir, './subcategories.yaml')
subcategories = YAML.load_file(subcategory_file)
orig_str = "subcategory: \"\""

subcategories.each do |category, files|
  files.each do |file|
    sub = "subcategory: \"#{category}\""
    ds = File.join(docs_dir, "data-sources", "#{file}.md")
    rs = File.join(docs_dir, "resources", "#{file}.md")
    if File.exist?(ds)
      content = File.read(ds)
      new_content = content.gsub(orig_str, sub)
      File.write(ds, new_content)
    end
    if File.exist?(rs)
      content = File.read(rs)
      new_content = content.gsub(orig_str, sub)
      File.write(rs, new_content)
    end
  end
end
