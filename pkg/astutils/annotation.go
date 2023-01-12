package astutils

type (
	Annotation    string
	AnnotationKey string
)

func (a Annotation) String() string { return string(a) }

const (
	// CommentProxy for struct while comment @Proxy generate a file with _gen.go suffix
	CommentProxy = Annotation("@Proxy")
	// CommentPointcut for struct function while comment @Pointcut generate a proxy func for proxy struct
	CommentPointcut = Annotation("@Pointcut")
	// CommentAspect for struct while comment @Aspect then use to enhance other function
	CommentAspect = Annotation("@Aspect")
	// CommentAdviceBefore for struct function while comment @Before then use to enhance other function
	CommentAdviceBefore = Annotation("@Before")
	// CommentAdviceAfter for struct function while comment @After then use to enhance other function
	CommentAdviceAfter = Annotation("@After")
	// CommentAdviceAround for struct function while comment @Around then use to enhance other function
	CommentAdviceAround = Annotation("@Around")
	// CommentComponent for struct factory method while comment @Component then use to inject proxy struct
	CommentComponent = Annotation("@Component")
	// CommentInject for struct field while comment @Inject then use to inject proxy struct
	CommentInject = Annotation("@Inject")

	// CommentKeyDefault key for comment params separated by "="
	CommentKeyDefault = AnnotationKey("default")
	// CommentKeyDefault abstract key for @Proxy comment
	CommentKeyAbstract = AnnotationKey("abstract")
	CommentKeySuffix   = AnnotationKey("suffix")
	CommentKeyCustom   = AnnotationKey("custom")
)

var (
	adviceAnnotationList = []Annotation{CommentAdviceBefore, CommentAdviceAfter, CommentAdviceAround}
	allAnnotationKey     = map[AnnotationKey]struct{}{
		CommentKeyDefault:  {},
		CommentKeyAbstract: {},
		CommentKeySuffix:   {},
		CommentKeyCustom:   {},
	}
	systemAnnotation = map[Annotation]struct{}{
		CommentProxy:        {},
		CommentPointcut:     {},
		CommentAspect:       {},
		CommentAdviceBefore: {},
		CommentAdviceAfter:  {},
		CommentAdviceAround: {},
		CommentComponent:    {},
		CommentInject:       {},
	}
)

func IsSystemAnnotation(anno Annotation) bool {
	_, ok := systemAnnotation[anno]
	return ok
}

func AdviceAnnotationList() []Annotation {
	return adviceAnnotationList
}

func IsSystemAnnotationKey(key AnnotationKey) bool {
	_, ok := allAnnotationKey[key]
	return ok
}
