require 'yajl'
require 'markdown_html'

module FlynnDoc
  class APIExample
    def self.compile(data, options = {})
      new(data, options).compile
    end

    attr_reader :data, :options

    def initialize(data, options = {})
      @data = data

      @options = {
        :path => path,
        :keyword => keyword
      }.merge(options)
    end

    def compile
      data.gsub(/\{(\w+) #{@options[:keyword]}\}/) do
        if api_examples.has_key?($1)
          Redcarpet::Markdown.new(Middleman::Renderers::MiddlemanRedcarpetHTML, REDCARPET_EXTENTIONS).render(api_examples[$1])
        else
          "#{$1} not found"
        end
      end
    end

    private

    def api_examples
      @api_examples ||= Yajl::Parser.parse(
        File.read(options[:path])
      )
    end

    def path
    end

    def keyword
      "example".freeze
    end
  end

  class ControllerAPIExample < APIExample
    private

    def path
      File.expand_path("../../../source/docs/controller_examples.json", __FILE__)
    end

    def keyword
      "controller_example".freeze
    end
  end
end
