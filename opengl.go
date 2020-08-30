package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func makeVao(points []float32, indices []uint32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)

	// VertexAttribPointer index refers to `layout (location = 0) ` in the vertex shader
	// stride can be set to 0 when the values are tightly packed
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(3*4), nil)
	gl.EnableVertexAttribArray(0)

	// unbind objects
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return vao
}

func draw(vao uint32, window *glfw.Window, program uint32) {
	processInput(window)

	// clear buffers
	gl.ClearColor(0.2, 0.3, 0.3, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// actual drawing
	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

	// Unbinding is optional if we always bind a VAO before a draw call
	// Also, would like to benchmark this
	// It is still safer to unbind so that if someone tries
	// to draw without binding a VAO prior, it fails right away
	gl.BindVertexArray(0)

	glfw.PollEvents()
	window.SwapBuffers()
}

func processInput(w *glfw.Window) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
}

func createWindow(width, height int, title string) (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.True)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create window: %s", err)
	}
	return window, nil
}

func createProgram(vsSource, fsSource string) (uint32, error) {
	vertexShader, err := compileShader(vsSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("could not compile vertex shader: %s", err)
	}
	fragmentShader, err := compileShader(fsSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, fmt.Errorf("could not compile fragment shader: %s", err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	gl.ValidateProgram(prog)

	// free memory once attached to a program
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// dont do this if we need to debug
	// the shaders in the GPU
	gl.DetachShader(prog, vertexShader)
	gl.DetachShader(prog, fragmentShader)

	if err := retrieveProgramLinkError(prog); err != nil {
		return 0, err
	}
	return prog, nil
}

func retrieveProgramLinkError(program uint32) error {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return fmt.Errorf("failed to link program: %v", log)
	}
	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()

	gl.CompileShader(shader)
	if err := retrieveShaderCompileError(shader); err != nil {
		return 0, err
	}
	return shader, nil
}

func retrieveShaderCompileError(shader uint32) error {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return fmt.Errorf("failed to compile shader: %v", log)
	}
	return nil
}