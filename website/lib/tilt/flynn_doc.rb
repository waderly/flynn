require 'tilt/template'
require 'flynn_doc'

module Tilt
  class FlynnDocTemplate < Template
    def evaluate(scope, locals, &block)
      ::FlynnDoc.compile(data)
    end
  end
end

